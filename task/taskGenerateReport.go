package task

import (
	"errors"
	"fmt"
	"time"
)

// GenerateReport simulates generating a report
type GenerateReport struct {
	Task
	Notify     []string `json:"notify"`
	ReportType string   `json:"reportType" json:"reportType"`
}

func (t *GenerateReport) ProcessTask() error {
	time.Sleep(t.MockProcessingTime)
	fmt.Printf("Report : %s generated, notifying %s", t.ReportType, t.Notify)
	t.Status = t.MockProcessingResult
	return nil
}

func (t GenerateReport) ValidateTask() error {
	if t.ReportType == "" {
		return errors.New("unsupported report type")
	}
	return nil
}
