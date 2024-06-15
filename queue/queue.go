package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type option func(q *Queue)

type Queue struct {
	mainTaskChan   *chan []*task.Task
	resultChan     chan []*task.Task
	ctx            context.Context
	maxQueueSize   int
	workerPoolSize int
	workerPool     *WorkerPool
	priorityChans  []chan task.Task
}

// CreateQueue creates and returns the Queue with predefined options
func CreateQueue(ctx context.Context, opts ...option) (*Queue, error) {
	q := Queue{
		ctx:            ctx,
		maxQueueSize:   100,
		workerPoolSize: 5,
		priorityChans:  make([]chan task.Task, 0, 2),
	}

	for _, opt := range opts {
		opt(&q)
	}

	if q.mainTaskChan == nil {
		return nil, fmt.Errorf("main task channel must be set")
	}

	for i := 1; i <= 2; i++ {
		q.priorityChans = append(q.priorityChans, make(chan task.Task))
	}

	q.workerPool = CreateWorkerPool(q.workerPoolSize, q.priorityChans...)

	return &q, nil
}

// WithMainQueue *required* the queue will listen to new tasks on this chan
func WithMainQueue(taskChan *chan []*task.Task) option {
	return func(q *Queue) {
		q.mainTaskChan = taskChan
	}
}

// WithMaxQueueSize the max queue size before it must be processed to avoid starvation
func WithMaxQueueSize(size int) option {
	return func(q *Queue) {
		q.maxQueueSize = size
	}
}

// WithMaxWorkerPoolSize the amount of workers in the worker pool
func WithMaxWorkerPoolSize(size int) option {
	return func(q *Queue) {
		q.workerPoolSize = size
	}
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {

	// Start listening to enqueue tasks
	go func() {
		slog.Info("queue has started listening")
		for {
			select {
			case <-q.ctx.Done():
				slog.Info("queue shutdown initiated, main context cancelled")
				return
			case tasks, ok := <-*q.mainTaskChan:
				if !ok {
					slog.Error("reading from empty channel")
					return
				}
				q.enqueue(tasks)
			}
		}
	}()

	// Start listening for task results
}

func (q *Queue) enqueue(tasks []*task.Task) error {

	for _, t := range tasks {
		t.Status = task.ProcessingEnqueued
		switch t.Priority {
		case task.High:
			q.priorityChans[0] <- *t
		case task.Low:
			q.priorityChans[1] <- *t
		default:
			q.priorityChans[len(q.priorityChans)] <- *t
		}

	}

	return nil
}

func (q *Queue) receiveResult() {

}
