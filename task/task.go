package task

import "fmt"

// TODO rename this
type Processing interface {
	ProcessTask() error
}

// TypeOf  enum describes supported type of report
type TypeOf string

const (
	TypeSendEmail      TypeOf = "SendEmail"
	TypeGenerateReport TypeOf = "GenerateReport"
	TypeCPUProcess     TypeOf = "CPUProcess"
)

// ExecutionPriority  enum describes execution priority of the task
type ExecutionPriority int

const (
	Low ExecutionPriority = iota
	High
)

// ProcessingStatus enum describes current state of the task
type CurrentStatus string

// TODO Maybe simplify these ?
const (
	ProcessingAwaiting      CurrentStatus = "Awaiting enqueue"
	ProcessingEnqueued      CurrentStatus = "Enqueued, awaiting processing"
	ProcessingSuccess       CurrentStatus = " Processed successfully"
	ProcessingAwaitingRetry CurrentStatus = "Awaiting retry"
	ProcessingFailed        CurrentStatus = "Failed to process"
)

type Task struct {
	Priority ExecutionPriority
	Type     TypeOf
	Status   CurrentStatus
}

type option func(task *Task)

func CreateTask(opts ...option) (*Task, error) {

	t := &Task{
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
