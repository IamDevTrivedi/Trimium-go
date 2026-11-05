// Package queue is the producer-facing wall over the background-task
// transport. Services receive a Queue and call domain verbs such as
// SendEmail; they never see asynq, payload marshaling, or enqueue
// mechanics. The consumer side (workers, task server) deliberately
// keeps full transport knowledge.
package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"server/internal/tasks"
)

// Queue is the contract handed to every service that needs to push
// background work. Implementations guarantee that a nil error means
// the payload was accepted for delivery; execution, retries and any
// queuing details stay behind this wall and are not the caller's
// concern.
type Queue interface {
	SendEmail(ctx context.Context, payload tasks.EmailPayload) error
}

type asynqQueue struct {
	client *asynq.Client
}

// New hides the compositional root's asynq client behind the Queue
// wall. The caller keeps ownership of the client, including closing
// it; Queue deliberately exposes no lifecycle.
func New(client *asynq.Client) Queue {
	return &asynqQueue{client: client}
}

// SendEmail enqueues an email:send task. The context bounds the
// enqueue operation only, never the execution of the delivered email.
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
