package queue

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/sinderpl/AsyncTaskProcessor/storage"
	"github.com/sinderpl/AsyncTaskProcessor/task"
)

// Package queue deals with receiving tasks, sending them to the worker pool and handling retries / backoff

type option func(q *Queue)

// Queue represents our queue handler taking care of all the retries, awaits and passing the tasks onto workers
type Queue struct {
	ctx            context.Context
	maxBufferSize  int
	workerPoolSize int
	maxTaskRetry   int
	db             storage.Storage

	mainTaskChan  *chan []*task.Task // we receive any new tasks on this channel
	resultChan    chan task.Task     // the workers can write the task status back to this channel
	priorityChans []chan task.Task   // deals with the different priorities low / high
	workerPool    *WorkerPool        // instance of our workers that are created here, could perhaps be externalised to its own package

	awaitingQueue linkedList // stores tasks when the buffered chans dont have capacity yet
}

// CreateQueue creates and returns the Queue with predefined options
func CreateQueue(ctx context.Context, opts ...option) (*Queue, error) {
	q := Queue{
		ctx:            ctx,
		maxBufferSize:  10,
		workerPoolSize: 5,
		maxTaskRetry:   0,

		priorityChans: make([]chan task.Task, 0, 2),
		resultChan:    make(chan task.Task),
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

	q.workerPool = CreateWorkerPool(q.workerPoolSize, q.resultChan, q.priorityChans)

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

// WithStorage adds persistent data store
func WithStorage(storage storage.Storage) option {
	return func(q *Queue) {
		q.db = storage
	}
}

// Start the queue starts listening to new tasks coming in
func (q *Queue) Start() {
	go q.awaitTasks()
	go q.awaitResults()
	go q.pushToProcess()
}

// pushToProcess scans the current queue and pushes to the workers when there is space in the buffered chans
// and a tasks backoff is finished or nil
func (q *Queue) pushToProcess() {
	for {
		// TODO add context deadline

		// Check if any of the chans have space for new tasks starting from highest priority one
		for priorityId := len(q.priorityChans) - 1; priorityId >= 0; priorityId-- {
			space := q.maxBufferSize - len(q.priorityChans[priorityId])

			if space > 0 {
				if q.awaitingQueue.getFirst() != nil {
					currNode := q.awaitingQueue.first
					for currNode != nil && space >= 0 {
						// Small hack but with having more time I would probably rewrite priorityChans to be a map
						// of taskId:taskPriority to allow for better prioritisation
						if int(currNode.t.Priority) == priorityId && (currNode.t.BackOffUntil == nil ||
							(currNode.t.BackOffUntil != nil && currNode.t.BackOffUntil.Before(time.Now()))) {

							// Enqueue the task to channel to be picked up by worker
							currNode.t.Status = task.ProcessingEnqueued
							slog.Info(fmt.Sprintf("enqueing task %s", currNode.t.Id))
							q.priorityChans[priorityId] <- *currNode.t
							q.awaitingQueue.pop(currNode)

							err := q.db.UpdateTask(currNode.t)
							if err != nil {
								slog.Error(fmt.Sprintf("failed to update task details to database: %v \n", err))
							}

							// Decrease current available space on the chan and pop the node from the awaiting queue
							space--
						}
						currNode = currNode.next
					}
				}
			}
		}
	}
}

// enqueue adds tasks to the awaiting channel queue to be processed when a worker is available
func (q *Queue) enqueue(tasks ...*task.Task) {
	for _, t := range tasks {
		q.awaitingQueue.append(t)
	}
}

// awaitTasks waits for new tasks to come in and push them to be processed
func (q *Queue) awaitTasks() {
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
}

// awaitResults waiting for task results so that it can retry or fail them
func (q *Queue) awaitResults() {
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
				if t.Retries >= q.maxTaskRetry {
					t.Status = task.ProcessingFailed
					fmt.Println(t.ErrorDetails)
					slog.Error(fmt.Sprintf("error while processing task: %s no retries left saving failed status, error: %v \n", t.Id, t.Error))
					err := q.db.UpdateTask(&t)
					if err != nil {
						slog.Error(fmt.Sprintf("failed to update task details to database: %v \n", err))
					}
					continue
				}

				t.Retries++
				if t.BackOffDuration != nil {
					bckOffUntil := time.Now().Add(*t.BackOffDuration)
					t.BackOffUntil = &bckOffUntil
				}
				slog.Info(fmt.Sprintf("failed: error while processing task: %s, retrying. retry attempt:%d error: %v \n", t.Id, t.Retries, t.Error))

				t.Error = nil
				q.enqueue(&t)
				continue
			}

			t.Status = task.ProcessingSuccess
			currTime := time.Now().UTC()
			t.FinishedAt = &currTime
			err := q.db.UpdateTask(&t)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to update task details to database: %v \n", err))
			}
			slog.Info(fmt.Sprintf("task:%s processed succesfully \n", t.Id))
		}
	}
}

// I thought a linked list would be good to keep track of execution
// This is due to channels in Go having to be buffered in order to keep items on their queue
// A unbuffered channel does not wait until we have a worker available to read it
// A buffered channel can only hold X tasks but will hold those until the workers pick them up which leaves us with
// in-between tasks that are not ready for the channels but still need to be scheduled
// In addition I wanted to have somewhere to store the tasks with backoff so they can be picked up later
// (editing an array mid iteration is not a good idea so thought linked list might work better here)
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

func (l *linkedList) getFirst() *task.Task {
	// TODO https://go.dev/doc/articles/race_detector#Runtime_Overheads
	l.listMutex.Lock()
	defer l.listMutex.Unlock()

	if l.first == nil {
		return nil
	}

	return l.first.t
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
