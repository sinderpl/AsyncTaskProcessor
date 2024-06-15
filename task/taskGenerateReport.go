package task

import (
	"errors"
	"fmt"
)

// GenerateReport simulates generating a report
type GenerateReport struct {
	Notify     []string `json:"notify"`
	ReportType string   `json:"reportType" json:"reportType"`
}

func (t *GenerateReport) ProcessTask() error {
	fmt.Printf("Report : %s generated, notifying %s \n", t.ReportType, t.Notify)
	return nil
}

func (t GenerateReport) ValidateTask() error {
	if t.ReportType == "" {
		return errors.New("unsupported report type")
	}
	return nil
}
