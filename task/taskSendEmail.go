package task

type SendEmailTask struct {
	Task
	SendTo   []string `json:"sendTo"`
	SendFrom string   `json:"sendFrom"`
	Title    string   `json:"title"`
	Body     string   `json:"body"`
}

func (t *SendEmailTask) ProcessTask() error {

	//TODO implement me
	panic("implement me")
}
