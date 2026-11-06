package workers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"server/internal/logger"
	"server/internal/mailer"
	"server/internal/tasks"
)

type EmailWorker struct {
	mailer *mailer.Client
}

func NewEmailWorker(m *mailer.Client) *EmailWorker {
	return &EmailWorker{mailer: m}
}

func (w *EmailWorker) HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.ParseEmailPayload(t)
	if err != nil {
		return fmt.Errorf("parse email task: %w: %w", err, asynq.SkipRetry)
	}

	log := logger.Get()

	if err := w.mailer.Send(ctx, payload.FromEmail, payload.ToEmail, payload.Subject, payload.HTMLBody); err != nil {
		return fmt.Errorf("send email to %v: %w", payload.ToEmail, err)
	}

	log.Info("Email sent successfully", "to", payload.ToEmail, "subject", payload.Subject)
	return nil
}
