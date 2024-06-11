package task

import "time"

// GenerateReport simulates generating a report
type GenerateReport struct {
	Task
	Notify     []string
	ReportType string `json:"sendTo"`
}

func (t *GenerateReport) ProcessTask() error {
	time.Sleep(time.Second * 5)

	return nil
}
