package task

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Processor interface {
	ProcessTask() error
}

// TypeOf enum describing supported type of report
type TypeOf string

const (
	TypeSendEmail      TypeOf = "SendEmail"
	TypeGenerateReport TypeOf = "GenerateReport"
	TypeCPUProcess     TypeOf = "CPUProcess"
)

// ExecutionPriority enum describing execution priority of the task
type ExecutionPriority int

const (
	Low ExecutionPriority = iota
	High
)

// CurrentStatus enum describing current state of the task
type CurrentStatus string

// TODO simplify ?
const (
	ProcessingAwaiting      CurrentStatus = "Awaiting enqueue"
	ProcessingEnqueued      CurrentStatus = "Enqueued, awaiting processing"
	ProcessingSuccess       CurrentStatus = " Processed successfully"
	ProcessingAwaitingRetry CurrentStatus = "Awaiting retry"
	ProcessingFailed        CurrentStatus = "Failed to process"
)

type Task struct {
	Id         string
	Priority   ExecutionPriority
	Type       TypeOf
	Status     CurrentStatus
	Error      *error
	CreatedAt  time.Time
	CreatedBy  string
	StartedAt  time.Time
	FinishedAt time.Time

	// MOCKING Just for simulating running not for prod
	MockProcessingTime   time.Duration
	MockProcessingResult CurrentStatus
}

type option func(task *Task)

// WithPriority sets tasks processing priority
func WithPriority(priority ExecutionPriority) option {
	return func(t *Task) {
		t.Priority = priority
	}
}

// WithType *Required* sets task type
func WithType(typeOf TypeOf) option {
	return func(t *Task) {
		t.Type = typeOf
	}
}

// WithCreatedBy *Required* sets created by user id
func WithCreatedBy(id string) option {
	return func(t *Task) {
		t.CreatedBy = id
	}
}

// CreateTask creates and validates a task with the supplied options
func CreateTask(opts ...option) (*Task, error) {

	t := &Task{
		Id:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
		Priority:  Low,
		Status:    ProcessingAwaiting,
	}

	fmt.Println(t.CreatedAt)

	for _, opt := range opts {
		opt(t)
	}

	if err := t.validateTask(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Task) validateTask() error {
	if t.Type == "" {
		return fmt.Errorf("task type must be set")
	}

	if t.CreatedBy == "" {
		return fmt.Errorf("creator user id must be set")
	}

	return nil
}

// MOCKING Execution

// WithMockProcessingTime *Required* simulates processing time
func WithMockProcessingTime(ti time.Duration) option {
	return func(t *Task) {
		t.MockProcessingTime = ti
	}
}

// WithMockProcessingResult *Required* simulates result
func WithMockProcessingResult(exp CurrentStatus) option {
	return func(t *Task) {
		t.MockProcessingResult = exp
	}
}
