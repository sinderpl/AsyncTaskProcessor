package queue

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sinderpl/AsyncTaskProcessor/task"
	"log/slog"
	"time"
)

type worker struct {
	Id string
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
func (w worker) Start(resultChan chan task.Task, workerChan []chan task.Task) {
	go func() {
		for {
			// TODO improve different channel prioritisation
			select {
			case t := <-workerChan[1]:
				w.process(t, resultChan)
			case t := <-workerChan[0]:
				w.process(t, resultChan)
			default:
				time.Sleep(100 * time.Millisecond) // Avoid busy-wait, poll every so often for new work
			}
		}
	}()
}

// process starts the task processing implementation and returns any errors
func (w worker) process(t task.Task, resultChan chan task.Task) {
	t.Status = task.Processing
	t.StartedAt = time.Now().UTC()
	slog.Info(fmt.Sprintf("worker %s is processing task: %s with priority: %d \n", w.Id, t.Id, t.Priority))
	err := t.ProcessableTask.ProcessTask()
	if err != nil {
		t.Status = task.ProcessingAwaitingRetry
		t.Error = err
		resultChan <- t
	}
}

// CreateWorkerPool initializes a new worker pool of size numWorkers and registers them to listen to 2 chans
func CreateWorkerPool(numWorkers int, resultChan chan task.Task, workChans []chan task.Task) *WorkerPool {

	pool := &WorkerPool{}

	for i := 1; i <= numWorkers; i++ {
		worker := createWorker()
		worker.Start(resultChan, workChans)
	}

	return pool
}

// RegisterNewChan adds a new queue for the workers to listen to
func (*WorkerPool) RegisterNewChan(newChan <-chan task.Task) {

}
