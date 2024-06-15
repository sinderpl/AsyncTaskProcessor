package queue

import (
	"fmt"
	"log/slog"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type option func(q *Queue)

type Queue struct {
	taskChan *chan []task.Task
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {
	go func() {
		slog.Info("queue has started listening")
		for {
			select {
			case tasks, ok := <-*q.taskChan:
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

// WithQueue *required* the queue will listen to new tasks on this chan
func WithQueue(taskChan *chan []task.Task) option {
	return func(q *Queue) {
		q.taskChan = taskChan
	}
}

func (q *Queue) enqueue(tasks []task.Task) error {

	for _, t := range tasks {
		fmt.Printf("enqueing: %v", t)
	}

	return nil
}
