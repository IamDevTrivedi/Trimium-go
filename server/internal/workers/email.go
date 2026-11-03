package workers

import (
	"context"
	"fmt"
	"server/internal/logger"
	"server/internal/tasks"

	"github.com/hibiken/asynq"
)

func HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.ParseEmailPayload(t)
	if err != nil {
		return fmt.Errorf("parse email task: %v: %w", err, asynq.SkipRetry)
	}

	log := logger.Get()
	log.Info("Sending email (DUMMY)", "to", payload.ToEmail, "subject", payload.Subject)
	return nil
}
