package queue

import (
	"fmt"
	"runtime"
)

type QueueItem interface {
	ExecuteTask() error
}

var queue chan QueueItem

func AddToQueue(item QueueItem) {
	queue <- item
}

func executor() {
	for {
		qi := <-queue
		err := qi.ExecuteTask()
		if err != nil {
			fmt.Printf("Error executing task for %+v", qi)
		}
	}
}

func init() {
	queue = make(chan QueueItem, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go executor()
	}

}
