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
	fmt.Printf("CPU process: %s completed \n", t.ProcessType)
	return nil
}

func (t *CPUProcess) ValidateTask() error {
	if t.ProcessType == "" {
		return errors.New("process type can't be empty")
	}
	return nil
}
