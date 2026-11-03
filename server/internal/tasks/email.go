package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeEmailSend = "email:send"

type EmailPayload struct {
	FromEmail string
	ToEmail   []string
	Subject   string
	HTMLBody  string
}

func NewEmailTask(payload *EmailPayload) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailSend, jsonPayload), nil
}

func ParseEmailPayload(t *asynq.Task) (*EmailPayload, error) {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
