package task

import (
	"errors"
	"fmt"
)

// CPUProcess simulates CPU processing time
type CPUProcess struct {
	ProcessType string `json:"processType"`
}

func (t *CPUProcess) ProcessTask() error {
	return fmt.Errorf("Error while processing task due to proces type failure: %s", t.ProcessType)
}

func (t *CPUProcess) ValidateTask() error {
	if t.ProcessType == "" {
		return errors.New("process type can't be empty")
	}
	return nil
}
