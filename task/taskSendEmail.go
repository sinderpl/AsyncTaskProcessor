package task

import (
	"errors"
	"fmt"
)

// SendEmail simulates sending a email
type SendEmail struct {
	SendTo   []string `json:"sendTo"`
	SendFrom string   `json:"sendFrom"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

func (t *SendEmail) ProcessTask() error {
	fmt.Printf("Email sent from : %s to : %s , subject: %s", t.SendFrom, t.SendTo, t.Subject)
	return nil
}

func (t SendEmail) ValidateTask() error {
	if len(t.SendTo) <= 0 {
		return errors.New("recipients cant be empty")
	}

	if t.SendFrom == "" {
		return errors.New("sender cant be empty")
	}

	if t.Subject == "" {
		return errors.New("subject cant be empty")
	}

	return nil
}
