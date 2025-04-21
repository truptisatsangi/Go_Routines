package main

import (
	"fmt"
	"sync"
	"time"
)

// go routine lets you run your code parallely, make sync code asyc for better performance and speed
func printMessage(message string) {
	for i := 0; i < 10; i++ {
		fmt.Println(message)
		time.Sleep(1 * time.Second)
	}
}

// channels

func sendData(ch chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 10; i++ {
		ch <- i
	}
	close(ch)
}

func receiveData(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range ch {
		fmt.Println("Data received: ", data)
	}
}

func main() {
	// Go routine
	go printMessage("Hello from Goroutine")
	fmt.Println("Main thread")

	// wait for go routine to complete execution
	time.Sleep(11 * time.Second)
	fmt.Println("After sleep")

	// CHANNEL
	var wg sync.WaitGroup
	wg.Add(2)
	ch := make(chan int)
	go sendData(ch, &wg)
	go receiveData(ch, &wg)
	wg.Wait() // wait for both go routine to finish

	//  time.Sleep(2 * time.Second)

	// -------------------
	ch := make(chan int, 3)

	ch <- 1
	ch <- 3
	ch <- 2

	fmt.Println(<-ch)
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

