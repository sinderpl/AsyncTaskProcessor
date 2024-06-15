package task

import (
	"errors"
	"fmt"
	"time"
)

// CPUProcess simulates CPU processing time
type CPUProcess struct {
	Task
	ProcessType string `json:"processType"`
}

func (t CPUProcess) ProcessTask() error {
	time.Sleep(t.MockProcessingTime)
	fmt.Printf("CPU process: %s completed, took %s", t.ProcessType, t.MockProcessingTime)
	t.Status = t.MockProcessingResult
	return nil
}

func (t CPUProcess) ValidateTask() error {
	if t.ProcessType == "" {
		return errors.New("process type can't be empty")
	}
	return nil
}
