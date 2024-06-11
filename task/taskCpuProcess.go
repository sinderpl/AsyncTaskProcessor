package task

import (
	"fmt"
	"time"
)

// CPUProcess simulates CPU processing time
type CPUProcess struct {
	Task
	ProcessType string
}

func (t CPUProcess) ProcessTask() error {
	time.Sleep(t.MockProcessingTime)
	fmt.Printf("CPU process: %s completed, took %s", t.ProcessType, t.MockProcessingTime)
	t.Status = t.MockProcessingResult
	return nil
}
