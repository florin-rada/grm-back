package queue

import (
	"runtime"
)

type QueueItem func(...interface{})
type Queue chan QueueItem

type QueueManager struct {
	queue        Queue
	numWorkers   int
	maxQueueSize int
}

// NewQueueManager creates a new queue and starts the workers (based on the numWorkers)
func NewQueueManager(maxQueueSize int, numWorkers int) QueueManager {
	qm := QueueManager{
		queue:        newQueue(maxQueueSize),
		numWorkers:   numWorkers,
		maxQueueSize: maxQueueSize,
	}
	qm.startWorkers()
	return qm
}

func (qm QueueManager) AddToQueue(item QueueItem) {
	qm.queue <- item
}

func (qm *QueueManager) AddWorkers(numWorkersToAdd int) {
	if numWorkersToAdd <= 0 {
		return
	}
	for i := 0; i < qm.numWorkers; i++ {
		go executor(qm.queue)
	}
	qm.numWorkers += numWorkersToAdd
}

func (qm *QueueManager) ShutDownSomeWorkers(numWorkers int) {
	// we make sure we don't add more shutdown signals than there are current workers because
	// otherwise when adding new workers, they will also shutdown until the total number of
	// shutdown signals are processed
	if numWorkers > qm.numWorkers {
		numWorkers = qm.numWorkers
	}
	for i := 0; i < numWorkers; i++ {
		qm.AddToQueue(nil)
	}
	qm.numWorkers -= numWorkers
}

func (qm QueueManager) startWorkers() {
	for i := 0; i < qm.numWorkers; i++ {
		go executor(qm.queue)
	}
}

// This calls the closure we received as argument when calling AddToQueu
func (qi QueueItem) ExecuteTask() {
	qi()
}

func newQueue(maxSize int) Queue {
	return make(Queue, maxSize)
}

var queue chan QueueItem

// replaced by qm.AddToQueue
// Receives a closure with signature func() and adds it to the queue
/* func AddToQueue(item QueueItem) {
	queue <- item
} */

// This is our executor, it loops and waits for tasks to do
// Runs in a spearate goroutine
func executor(queue Queue) {
	for {
		qi := <-queue
		// We exit worker if we receive nil, this is our exit signal
		if qi == nil {
			return
		}
		qi.ExecuteTask()

	}
}

func init() {
	queue = make(chan QueueItem, runtime.NumCPU()*1000)
	for i := 0; i < runtime.NumCPU()*10; i++ {
		go executor(queue)
	}

}
