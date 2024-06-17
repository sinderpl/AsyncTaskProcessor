package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Processable interface {
	ProcessTask() error
	ValidateTask() error
}

// TypeOf enum describing supported type of report
type TypeOf string

// TODO perhaps find a better way to extend these enums
// If updating make sure to update the isValidTypeOf function
const (
	TypeSendEmail      TypeOf = "SendEmail"
	TypeGenerateReport TypeOf = "GenerateReport"
	TypeCPUProcess     TypeOf = "CPUProcess"
)

func isValidTypeOf(typeOf TypeOf) bool {
	// Unfortunate workaround due to lack of enums in GO
	switch typeOf {
	case TypeSendEmail, TypeGenerateReport, TypeCPUProcess:
		return true
	}
	return false
}

// ExecutionPriority enum describing execution priority of the task
type ExecutionPriority int

const (
	Low ExecutionPriority = iota
	High
)

// CurrentStatus enum describing current state of the task
type CurrentStatus string

const (
	ProcessingAwaiting      CurrentStatus = "Awaiting enqueue"
	ProcessingEnqueued      CurrentStatus = "Enqueued, awaiting processing"
	Processing              CurrentStatus = "Being processed by worker"
	ProcessingSuccess       CurrentStatus = "Processed successfully"
	ProcessingAwaitingRetry CurrentStatus = "Awaiting retry"
	ProcessingFailed        CurrentStatus = "Failed to process"
)

type Task struct {
	Id              string
	Priority        ExecutionPriority
	TaskType        TypeOf
	Status          CurrentStatus
	BackOffDuration *time.Duration
	Payload         json.RawMessage
	ProcessableTask Processable

	CreatedAt time.Time
	CreatedBy string

	StartedAt  *time.Time
	FinishedAt *time.Time

	Retries      int
	BackOffUntil *time.Time
	Error        error
	ErrorDetails string // Used for DB persistence
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
		t.TaskType = typeOf
	}
}

// WithCreatedBy *Required* sets created by user id
func WithCreatedBy(id string) option {
	return func(t *Task) {
		t.CreatedBy = id
	}
}

// WithBackoffTime sets created by user id
func WithBackoffTime(backoffDuration string) option {
	return func(t *Task) {
		if backoffDuration != "" {
			h, err := time.ParseDuration(backoffDuration)
			if err != nil {
				return
			}
			t.BackOffDuration = &h
		}
	}
}

// WithPayload sets created by user id
func WithPayload(payload json.RawMessage) option {
	return func(t *Task) {
		t.Payload = payload
	}
}

// CreateTask creates and validates a task with the supplied options and converts the data to the concrete implementation
func CreateTask(opts ...option) (*Task, error) {

	t := &Task{
		Id:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
		Priority:  Low,
		Status:    ProcessingAwaiting,
	}

	for _, opt := range opts {
		opt(t)
	}

	if err := t.validateTask(); err != nil {
		return nil, fmt.Errorf("task validation failed: %v", err)
	}

	process, err := t.ParseTaskType()
	if err != nil {
		return nil, fmt.Errorf("task parsing failed: %v", err)
	}

	if err := process.ValidateTask(); err != nil {
		return nil, fmt.Errorf(" task payload validation failed: %v", err)
	}

	// TODO The composition here could be improved
	// Assign parsed task and remove the raw payload
	t.ProcessableTask = process

	return t, nil
}

func (t *Task) validateTask() error {
	if t.TaskType == "" {
		return fmt.Errorf("task type must be set")
	}

	if !isValidTypeOf(t.TaskType) {
		return fmt.Errorf("unsupported task type")
	}

	if t.CreatedBy == "" {
		return fmt.Errorf("creator user id must be set")
	}

	if t.Priority < Low || t.Priority > High {
		return fmt.Errorf("unsupported priority")
	}

	return nil
}

// ParseTaskType parses the task payload into the correct type which implements the Processable interface
func (t *Task) ParseTaskType() (Processable, error) {
	var payload interface{}

	switch t.TaskType {
	case TypeSendEmail:
		payload = new(SendEmail)
	case TypeGenerateReport:
		payload = new(GenerateReport)
	case TypeCPUProcess:
		payload = new(CPUProcess)
	default:
		return nil, errors.New("unsupported data type")
	}

	err := json.Unmarshal(t.Payload, payload)
	if err != nil {
		return nil, errors.New("failed to unmarshal task data payload")
	}

	return payload.(Processable), nil
}
