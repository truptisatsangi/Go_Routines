// Build a robust Go application that implements Create, Read, Update, and Delete (CRUD) functionalities for managing data entries.
// The first goroutine will handle multiple data types, unmarshalling the data using Google Protobuf.
// This data will then be sent over a channel to the second goroutine, which listens and processes the data to create database entries.
// Implement proper context handling to gracefully close all goroutines when the application shuts down, ensuring no data loss or abrupt terminations.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
	"go-routine-assignment/protodata"
)

// Seperate struct for logic
type DataEntry struct {
	Id    string
	Name  string
	Value int32
}

func (d *DataEntry) Marshal() ([]byte, error) {
	return proto.Marshal(&protodata.DataProto{
		Id:    d.Id,
		Name:  d.Name,
		Value: d.Value,
	})
}

func UnmarshalData(data []byte) (*DataEntry, error) {
	var protoData protodata.DataProto

	// Unmarshal puts original data in protoData
	err := proto.Unmarshal(data, &protoData)

	if err != nil {
		return nil, err
	}
	return &DataEntry{
		Id:    protoData.Id,
		Name:  protoData.Name,
		Value: protoData.Value,
	}, nil
}

// In-memory store
var dbStore = make(map[string]*DataEntry)

var mu sync.RWMutex

// CRUD operations
func Create(entry *DataEntry) {
	mu.Lock()
	defer mu.Unlock()
	dbStore[entry.Id] = entry
	log.Println("Created:", entry)
}

func Read(id string) *DataEntry {
	mu.RLock()
	defer mu.RUnlock()
	return dbStore[id]
}

func Delete(id string) {
	mu.Lock()
	defer mu.Unlock()
	delete(dbStore, id)
	log.Println("Deleted entry with ID:", id)

}

func DataProducer(ctx context.Context, ch chan<- []byte) {
	defer close(ch)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Producer shutting down....")
			return
		case <-ticker.C:
			// Populating dummy entries like this id-1, Name-1
			entry := &DataEntry{
				Id:    fmt.Sprintf("id-%d", count),
				Name:  fmt.Sprintf("Name-%d", count),
				Value: int32(count),
			}
			data, err := entry.Marshal()
			if err != nil {
				log.Println("Marshal error:", err)
				continue
			}
			ch <- data
			count++
		}
	}
}

func DataConsumer(ctx context.Context, ch <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer shutting down...")
			return

		case data, ok := <-ch:
			if !ok {
				log.Println("Channel closed. Consumer exiting...")
				return
			}
			entry, err := UnmarshalData(data)
			if err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}
			Create(entry)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan []byte, 10)

	// Signal handling for graceful shutdown

	// 	Creates a channel to receive OS signals.
	// It’s buffered (1) to ensure at least one signal is caught without blocking.
	sigs := make(chan os.Signal, 1)
	
	// syscall.SIGINT: Sent when you press Ctrl+C
	// syscall.SIGTERM: Sent when the app is terminated (e.g. from a script, Kubernetes pod, etc.).
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // The line <-sigs blocks and waits until a signal is received.
		cancel()
	}()
	go DataProducer(ctx, ch)
	DataConsumer(ctx, ch)

	log.Println("Application shut down gracefully")

}
