package queue

import (
	"runtime"
)

type QueueItem interface {
	ExecuteTask()
}

var queue chan QueueItem

func AddToQueue(item QueueItem) {
	queue <- item
}

func executor() {
	for {
		qi := <-queue
		qi.ExecuteTask()

	}
}

func init() {
	queue = make(chan QueueItem, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go executor()
	}

}
