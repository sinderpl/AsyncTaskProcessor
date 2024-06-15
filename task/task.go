package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TODO rename ?
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
	ProcessingSuccess       CurrentStatus = "Processed successfully"
	ProcessingAwaitingRetry CurrentStatus = "Awaiting retry"
	ProcessingFailed        CurrentStatus = "Failed to process"
)

type Task struct {
	Id              string
	Priority        ExecutionPriority
	Type            TypeOf
	Status          CurrentStatus
	Payload         json.RawMessage
	ProcessableTask Processable

	CreatedAt time.Time
	CreatedBy string

	// TODO add a running log
	StartedAt  time.Time
	FinishedAt time.Time

	Retries int
	Error   *error
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

	process, err := t.parseTaskType()
	if err != nil {
		return nil, fmt.Errorf("task parsing failed: %v", err)
	}

	if err := process.ValidateTask(); err != nil {
		return nil, fmt.Errorf(" task payload validation failed: %v", err)
	}

	// TODO The composition here could be improved
	// Assign parsed task and remove the raw payload
	t.ProcessableTask = process
	t.Payload = nil

	return t, nil
}

func (t *Task) validateTask() error {
	if t.Type == "" {
		return fmt.Errorf("task type must be set")
	}

	if !isValidTypeOf(t.Type) {
		return fmt.Errorf("unsupported task type")
	}

	if t.CreatedBy == "" {
		return fmt.Errorf("creator user id must be set")
	}

	return nil
}

func (t *Task) parseTaskType() (Processable, error) {
	var payload interface{}

	switch t.Type {
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
