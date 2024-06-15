package queue

import (
	"fmt"
	"log/slog"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type option func(q *Queue)

type Queue struct {
	mainTaskChan *chan []task.Task
	maxQueueSize int32
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {
	go func() {
		slog.Info("queue has started listening")
		for {
			select {
			case tasks, ok := <-*q.mainTaskChan:
				if !ok {
					slog.Error("reading from empty chane")
					//panic("reading from empty chanel")
					return
				}
				q.enqueue(tasks)
			}
		}
	}()
}

// CreateQueue creates and returns the Queue with predefined options
func CreateQueue(opts ...option) *Queue {
	q := Queue{}

	for _, opt := range opts {
		opt(&q)
	}

	return &q
}

// WithMainQueue *required* the queue will listen to new tasks on this chan
func WithMainQueue(taskChan *chan []task.Task) option {
	return func(q *Queue) {
		q.mainTaskChan = taskChan
	}
}

// WithMaxQueueSize the max queue size before it must be processed to avoid starvation
func WithMaxQueueSize(size int32) option {
	return func(q *Queue) {
		q.maxQueueSize = size
	}
}

func (q *Queue) enqueue(tasks []task.Task) error {

	for _, t := range tasks {
		fmt.Printf("enqueing: %v", t)
		t.ProcessableTask.ProcessTask()
	}

	return nil
}
