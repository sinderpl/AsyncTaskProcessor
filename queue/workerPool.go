package queue

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sinderpl/AsyncTaskProcessor/task"
	"log/slog"
	"time"
)

type worker struct {
	Id           string
	IsProcessing bool
}

func createWorker() worker {
	return worker{
		Id: uuid.New().String(),
	}
}

type WorkerPool struct {
	TaskQueue chan task.Task
}

// Start starts the worker to process tasks from multiple channels.
func (w worker) Start(channels ...chan task.Task) {
	go func() {
		for {
			// TODO how to prioritise different channels
			select {
			case t := <-channels[0]:
				w.IsProcessing = true
				slog.Info(fmt.Sprintf("worker %s is processing task: %s with priority: %d \n", w.Id, t.Id, t.Priority))
				err := t.ProcessableTask.ProcessTask()
				if err != nil {
					t.Status = task.ProcessingAwaitingRetry
					w.IsProcessing = false

				}

				w.IsProcessing = false
			case task := <-channels[1]:
				w.IsProcessing = true
				slog.Info(fmt.Sprintf("worker %s is processing task: %s with priority: %d \n", w.Id, task.Id, task.Priority))
				err := task.ProcessableTask.ProcessTask()
				if err != nil {

				}
				w.IsProcessing = false
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (w worker) process(t *task.Task) {

}

func CreateWorkerPool(numWorkers int, chans ...chan task.Task) *WorkerPool {

	pool := &WorkerPool{}

	for i := 1; i <= numWorkers; i++ {
		worker := createWorker()
		worker.Start(chans...)
	}

	return pool
}

// RegisterNewChan adds a new queue for the workers to listen to
func (*WorkerPool) RegisterNewChan(newChan <-chan task.Task) {

}
