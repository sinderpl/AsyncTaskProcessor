package queue

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type option func(q *Queue)

type Queue struct {
	ctx            context.Context
	maxQueueSize   int
	workerPoolSize int
	maxTaskRetry   int

	mainTaskChan  *chan []*task.Task
	resultChan    chan *task.Task
	priorityChans []chan task.Task
	workerPool    *WorkerPool
}

// CreateQueue creates and returns the Queue with predefined options
func CreateQueue(ctx context.Context, opts ...option) (*Queue, error) {
	q := Queue{
		ctx:            ctx,
		maxQueueSize:   100,
		workerPoolSize: 5,
		maxTaskRetry:   1,
		priorityChans:  make([]chan task.Task, 0, 2),
		resultChan:     make(chan *task.Task),
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
		if taskChan == nil {
			log.Fatal("task listening chat must be set")
		}
		q.mainTaskChan = taskChan
	}
}

// WithMaxQueueSize the max queue size before it must be processed to avoid starvation
func WithMaxQueueSize(size int) option {
	return func(q *Queue) {
		if size > 1 {
			q.maxQueueSize = size
		}
	}
}

// WithMaxWorkerPoolSize the amount of workers in the worker pool
func WithMaxWorkerPoolSize(size int) option {
	return func(q *Queue) {
		if size > 1 {
			q.workerPoolSize = size
		}
	}
}

// WithMaxTaskRetry the amount of times a task will be retried
func WithMaxTaskRetry(retries int) option {
	return func(q *Queue) {
		if retries > 1 {
			q.maxTaskRetry = retries
		}
	}
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {
	go q.awaitTasks()
	go q.awaitResults()
}

// enqueue adds tasks to their respective priority queue to be processed when a worker is available
func (q *Queue) enqueue(tasks ...*task.Task) {

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
}

// awaitTasks goroutine waiting for new tasks coming in
func (q *Queue) awaitTasks() {
	go func() {
		slog.Info("await tasks queue has started listening")
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
				q.enqueue(tasks...)
			}
		}
	}()
}

// awaitResults goroutine waiting for task results so that it can retry or fail them
func (q *Queue) awaitResults() {
	go func() {
		slog.Info("await results queue has started listening")
		for {
			select {
			case <-q.ctx.Done():
				slog.Info("queue shutdown initiated, main context cancelled")
				return
			case t, ok := <-q.resultChan:
				if !ok {
					slog.Error("reading from empty channel")
					return
				}

				if t.Error != nil {
					if t.Retries > q.maxTaskRetry {
						t.Status = task.ProcessingFailed
						// TODO add to deadletter
						continue
					}
					t.Retries++
					slog.Error(fmt.Sprintf("error while processing task %s retryNum:%d error: %v \n", t.Id, t.Retries, t.Error))

					t.Error = nil
					q.enqueue(t)
				}

				t.Status = task.ProcessingSuccess
				t.FinishedAt = time.Now()

			}
		}
	}()
}
