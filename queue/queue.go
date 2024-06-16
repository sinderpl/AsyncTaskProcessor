package queue

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type option func(q *Queue)

type Queue struct {
	ctx            context.Context
	maxBufferSize  int
	workerPoolSize int
	maxTaskRetry   int

	mainTaskChan  *chan []*task.Task
	resultChan    chan *task.Task
	priorityChans []chan task.Task
	workerPool    *WorkerPool

	awaitingQueue   linkedList
	deadLetterQueue []*task.Task // TODO this needs management as it will grow forever
}

// CreateQueue creates and returns the Queue with predefined options
func CreateQueue(ctx context.Context, opts ...option) (*Queue, error) {
	q := Queue{
		ctx:            ctx,
		maxBufferSize:  10,
		workerPoolSize: 5,
		maxTaskRetry:   0,

		priorityChans:   make([]chan task.Task, 0, 2),
		resultChan:      make(chan *task.Task),
		deadLetterQueue: make([]*task.Task, 0),
		awaitingQueue: linkedList{
			listMutex: sync.Mutex{},
		},
	}

	for _, opt := range opts {
		opt(&q)
	}

	if q.mainTaskChan == nil {
		return nil, fmt.Errorf("main task channel must be set")
	}

	for i := 1; i <= 2; i++ {
		q.priorityChans = append(q.priorityChans, make(chan task.Task, q.maxBufferSize))
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

// WithMaxBufferSize the max queue size before it must be processed to avoid starvation
func WithMaxBufferSize(size int) option {
	return func(q *Queue) {
		if size > 1 {
			q.maxBufferSize = size
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
	go q.pushToProcess()
}

type linkedList struct {
	listMutex sync.Mutex
	first     *node
	last      *node
}

type node struct {
	t    *task.Task
	next *node
	prev *node
}

func (l *linkedList) append(t *task.Task) {
	l.listMutex.Lock()
	defer l.listMutex.Unlock()

	node := &node{
		t:    t,
		next: nil,
		prev: nil,
	}

	if l.first == nil {
		l.first = node
		l.last = node
		return
	}

	l.last.next = node
	node.prev = l.last
	l.last = node
}

func (l *linkedList) pop(n *node) {
	l.listMutex.Lock()
	defer l.listMutex.Unlock()

	if l.first == n {
		l.first = n.next
	}

	if l.last == n {
		l.last = n.prev
	}

	if n.prev != nil {
		n.prev.next = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	}
}

// pushToProcess scans the current queue and pushes to the workers when there is space in the buffered chans
// and a tasks backoff is finished or nil
func (q *Queue) pushToProcess() {
	go func() {
		for {

			// Check if any of the chans have space for new tasks starting from highest priority one
			for priorityId := len(q.priorityChans) - 1; priorityId >= 0; priorityId-- {
				space := q.maxBufferSize - len(q.priorityChans[priorityId])

				if space > 0 {
					if q.awaitingQueue.first != nil {
						currNode := q.awaitingQueue.first
						for currNode != nil && space >= 0 {
							// Small hack but with having more time I would probably rewrite priorityChans to be a map
							// of taskId:taskPriority to allow for better prioritisation
							if int(currNode.t.Priority) == priorityId && (currNode.t.BackOffUntil == nil ||
								(currNode.t.BackOffUntil != nil && currNode.t.BackOffUntil.After(time.Now()))) {

								// Enqueue the task to channel to be picked up by worker
								currNode.t.Status = task.ProcessingEnqueued
								slog.Info(fmt.Sprintf("enqueing task %s", currNode.t.Id))
								q.priorityChans[priorityId] <- *currNode.t
								q.awaitingQueue.pop(currNode)

								// Decrease current available space on the chan and pop the node from the awaiting queue
								space--
							}
							currNode = currNode.next
						}
					}
				}
			}
		}
	}()
}

// enqueue adds tasks to the awaiting channel queue to be processed when a worker is available
func (q *Queue) enqueue(tasks ...*task.Task) {
	for _, t := range tasks {
		q.awaitingQueue.append(t)
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
						slog.Error(fmt.Sprintf("error while processing task: %s no retries left, error: %v \n", t.Id, t.Error))
						q.deadLetterQueue = append(q.deadLetterQueue, t)
						continue
					}

					t.Retries++
					if t.BackOffDuration != nil {
						bckOffUntil := time.Now().Add(*t.BackOffDuration)
						t.BackOffUntil = &bckOffUntil
					}
					slog.Info(fmt.Sprintf("failed: error while processing task:%s, retrying. retryNum:%d error: %v \n", t.Id, t.Retries, t.Error))

					t.Error = nil
					q.enqueue(t)
					continue
				}

				t.Status = task.ProcessingSuccess
				t.FinishedAt = time.Now()
				slog.Info(fmt.Sprintf("task:%s processed succesfully \n", t.Id))
			}
		}
	}()
}
