package main

import (
	"back/pkg/models/queue"
	"runtime"
)

func loadData() error {

}

func main() {
	qm := queue.NewQueueManager(runtime.NumCPU()*1000, runtime.NumCPU()*100)

	/* replyChan := make(chan int, 3)
	min := int(0)
	max := int(100)
	numToGen := 10
	func(out chan int, min int, max int, toGen int) {
		qm.AddToQueue(func(...interface{}) {
			for i := 0; i < numToGen; i++ {
				num := rand.Intn(max)
				replyChan <- num
			}
			close(replyChan)
		})
	}(replyChan, min, max, numToGen)
	generatedNum := []int{}
	for i := range replyChan {
		generatedNum = append(generatedNum, i)
	}
	fmt.Printf("%+v", generatedNum) */
}
