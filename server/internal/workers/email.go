package workers

import (
	"context"
	"fmt"
	"server/internal/logger"
	"server/internal/mailer"
	"server/internal/tasks"

	"github.com/hibiken/asynq"
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
		return fmt.Errorf("parse email task: %v: %w", err, asynq.SkipRetry)
	}

	log := logger.Get()
	log.Info("Sending email", "to", payload.ToEmail, "subject", payload.Subject)

	if err := w.mailer.Send(ctx, payload.FromEmail, payload.ToEmail, payload.Subject, payload.HTMLBody); err != nil {
		return fmt.Errorf("send email to %v: %w", payload.ToEmail, err)
	}

	log.Info("Email sent successfully", "to", payload.ToEmail, "subject", payload.Subject)
	return nil
}
