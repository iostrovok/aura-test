package main

/*
	Application for simple test the service to create and list.
	HOST & URL are defined in ../helpers/helpers.go
*/

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iostrovok/aura-test/console/helpers"
)

// it sends create sessions threading * requestPerThread each delay timeout
const countCycles = 100
const delay = 500 * time.Millisecond
const threading = 10
const requestPerThread = 100

func main() {
	successCount := new(int32)
	errorCount := new(int32)

	count := countCycles
	for {
		count--
		oneSecondRequest(successCount, errorCount)
		helpers.List()

		if count <= 0 {
			break
		}
		<-time.After(delay)
	}

	fmt.Printf("successCount: %d\n", atomic.LoadInt32(successCount))
	fmt.Printf("errorCount: %d\n", atomic.LoadInt32(errorCount))
}

func oneSecondRequest(successCount, errorCount *int32) {
	wg := sync.WaitGroup{}
	startTime := time.Now()
	for i := 0; i < threading; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			createSession(successCount, errorCount)
		}()
	}

	wg.Wait()
	fmt.Printf("time: %+v\n", time.Now().Sub(startTime))
}

func createSession(successCount, errorCount *int32) {
	i := requestPerThread
	client := &http.Client{}

	for i > 0 {
		i--

		if helpers.CreateSession(client) != "" {
			atomic.AddInt32(successCount, 1)
		} else {
			atomic.AddInt32(errorCount, 1)
		}
	}
}
