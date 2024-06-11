package task

import (
	"fmt"
	"time"
)

// GenerateReport simulates generating a report
type GenerateReport struct {
	Task
	Notify     []string
	ReportType string `json:"sendTo"`
}

func (t *GenerateReport) ProcessTask() error {
	time.Sleep(t.MockProcessingTime)
	fmt.Printf("Report : %s generated, notifying %s", t.ReportType, t.Notify)
	t.Status = t.MockProcessingResult
	return nil
}
