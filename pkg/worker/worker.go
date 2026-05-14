package worker

import "github.com/gammazero/workerpool"

type WorkerPool struct {
	Worker *workerpool.WorkerPool
}

func NewWorker(numWorkers int) *WorkerPool {
	wp := workerpool.New(numWorkers)
	return &WorkerPool{
		Worker: wp,
	}
}

func (wp *WorkerPool) Stop() {
	wp.Worker.StopWait()
}
