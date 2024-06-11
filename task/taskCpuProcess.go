package task

import "time"

// CPUProcess simulates CPU processing time
type CPUProcess struct {
	Task
	sleepTime time.Duration
}

func (t CPUProcess) ProcessTask() error {
	time.Sleep(t.sleepTime)

	return nil
}
