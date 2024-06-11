package task

type GenerateReport struct {
	Task
	ReportType string `json:"sendTo"`
}

func (t *GenerateReport) ProcessTask() error {

	//TODO implement me
	panic("implement me")
}
