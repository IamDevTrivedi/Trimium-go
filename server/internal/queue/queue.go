package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"server/internal/tasks"
)

type Queue interface {
	SendEmail(ctx context.Context, payload tasks.EmailPayload) error
}

type asynqQueue struct {
	client *asynq.Client
}

func New(client *asynq.Client) Queue {
	return &asynqQueue{client: client}
}

func (q *asynqQueue) SendEmail(ctx context.Context, payload tasks.EmailPayload) error {
	task, err := tasks.NewEmailTask(&payload)
	if err != nil {
		return fmt.Errorf("queue: create email task: %w", err)
	}

	if _, err := q.client.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("queue: enqueue email task: %w", err)
	}

	return nil
}
