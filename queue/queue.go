package queue

import "github.com/sinderpl/AsyncTaskProcessor/task"

type option func(q *Queue)

type Queue struct {
	taskChan *chan []task.Task
}

func (q *Queue) Enqueue(t *task.Task) error {

	return nil
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {
	//go func() {
	//	for {
	//
	//	}
	//}()
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
