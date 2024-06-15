package queue

import (
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
			case task := <-channels[0]:
				w.IsProcessing = true
				slog.Info("worker %s is processing task: %s with priority: %d", w.Id, task.Id, task.Priority)
				err := task.ProcessableTask.ProcessTask()
				if err != nil {
					w.IsProcessing = false

				}

				w.IsProcessing = false
			case task := <-channels[1]:
				w.IsProcessing = true
				slog.Info("worker %s is processing task: %s with priority: %d", w.Id, task.Id, task.Priority)
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
