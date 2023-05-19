package queue

import (
	"runtime"
)

type QueueItem func(...interface{})

// This calls the closure we received as argument when calling AddToQueu
func (qi QueueItem) ExecuteTask() {
	qi()
}

var queue chan QueueItem

// Receives a closure with signature func() and adds it to the queue
func AddToQueue(item QueueItem) {
	queue <- item
}

// This is our executor, it loops and waits for tasks to do
// Runs in a spearate goroutine
func executor() {
	for {
		qi := <-queue
		qi.ExecuteTask()

	}
}

func init() {
	queue = make(chan QueueItem, runtime.NumCPU()*1000)
	for i := 0; i < runtime.NumCPU()*10; i++ {
		go executor()
	}

}
