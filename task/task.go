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
	ProcessingRejected      CurrentStatus = "Rejected"
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

	StartedAt  time.Time
	FinishedAt time.Time

	Error *error

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

	fmt.Println(t.CreatedAt)

	for _, opt := range opts {
		opt(t)
	}

	if err := t.validateTask(); err != nil {
		return nil, err
	}

	process, err := t.parseTaskType()

	fmt.Println(process)

	if err != nil {
		return nil, err
	}

	if err := process.ValidateTask(); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

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
