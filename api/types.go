package api

import (
	"encoding/json"
	"time"

	"github.com/sinderpl/AsyncTaskProcessor/task"
)

type EnqueueTaskPayload struct {
	Tasks []struct {
		TaskType        task.TypeOf            `json:"taskType"`
		Priority        task.ExecutionPriority `json:"priority,omitempty"`
		BackOffDuration *time.Duration         `json:"backOffDuration, omitempty"`
		Payload         json.RawMessage        `json:"payload,omitempty"`
	} `json:"Tasks"`
}

type EnqueueTaskResponse struct {
	Tasks  []TaskResponse `json:"tasks"`
	Status string         `json:"status"`
}

type TaskResponse struct {
	Id       string                 `json:"id"`
	TaskType task.TypeOf            `json:"taskType"`
	Priority task.ExecutionPriority `json:"priority"`
	Status   task.CurrentStatus     `json:"status"`
	Err      string                 `json:"err,omitempty"`
}
