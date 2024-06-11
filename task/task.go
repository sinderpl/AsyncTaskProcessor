package task

import (
	"fmt"
	"github.com/google/uuid"
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
	Id       string
	Priority ExecutionPriority
	Type     TypeOf
	Status   CurrentStatus
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

// CreateTask creates and validates a task with the supplied options
func CreateTask(opts ...option) (*Task, error) {

	t := &Task{
		Id:       uuid.New().String(),
		Priority: Low,
		Status:   ProcessingAwaiting,
	}

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
		return fmt.Errorf("task type must be specified")
	}

	return nil
}
