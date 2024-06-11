package task

import "time"

// SendEmail
type SendEmail struct {
	Task
	SendTo   []string `json:"sendTo"`
	SendFrom string   `json:"sendFrom"`
	Title    string   `json:"title"`
	Body     string   `json:"body"`
}

func (t *SendEmail) ProcessTask() error {

	time.Sleep(time.Second * 5)
	return nil
}
