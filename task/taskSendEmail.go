package task

import (
	"fmt"
)

// SendEmail simulates sending a email
type SendEmail struct {
	//Task
	SendTo   []string `json:"sendTo"`
	SendFrom string   `json:"sendFrom"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

func (t *SendEmail) ProcessTask() error {
	//time.Sleep(t.MockProcessingTime)
	fmt.Printf("Email sent from : %s to : %s , subject: %s", t.SendFrom, t.SendTo, t.Subject)
	//t.Status = t.MockProcessingResult
	return nil
}
