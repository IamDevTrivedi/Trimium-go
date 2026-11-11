package workers

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"server/internal/mailer"
	"server/internal/tasks"
)

// newTestMailer returns a mailer.Client pointed at a non-existent SMTP
// server. It is only used to satisfy NewEmailWorker's constructor and
// is never actually dialled in the parse-failure path.
func newTestMailer(t *testing.T) *mailer.Client {
	t.Helper()
	c, err := mailer.New(mailer.Options{
		Host: "127.0.0.1",
		Port: 1,
		TLS:  mailer.TLSNone,
	})
	if err != nil {
		t.Fatalf("mailer.New: %v", err)
	}
	return c
}

// TestNewEmailWorker verifies the constructor returns a non-nil worker.
func TestNewEmailWorker(t *testing.T) {
	w := NewEmailWorker(newTestMailer(t))
	if w == nil {
		t.Fatal("expected non-nil EmailWorker")
	}
}

// TestHandleEmailSend_InvalidPayload checks that a corrupted task
// payload is rejected immediately (SkipRetry wrapped) without
// attempting any network call.
func TestHandleEmailSend_InvalidPayload(t *testing.T) {
	w := NewEmailWorker(newTestMailer(t))

	badTask := asynq.NewTask(tasks.TypeEmailSend, []byte("not-json{{{"))
	err := w.HandleEmailSend(context.Background(), badTask)

	if err == nil {
		t.Fatal("expected an error for invalid payload, got nil")
	}

	// The handler must wrap asynq.SkipRetry so the queue does not retry.
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("expected error chain to contain asynq.SkipRetry, got: %v", err)
	}
}

// TestHandleEmailSend_EmptyPayload mirrors the invalid-payload case
// with a completely empty byte slice.
func TestHandleEmailSend_EmptyPayload(t *testing.T) {
	w := NewEmailWorker(newTestMailer(t))

	emptyTask := asynq.NewTask(tasks.TypeEmailSend, []byte(""))
	err := w.HandleEmailSend(context.Background(), emptyTask)

	if err == nil {
		t.Fatal("expected an error for empty payload, got nil")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("expected error chain to contain asynq.SkipRetry, got: %v", err)
	}
}

// TestHandleEmailSend_ValidPayload_DialError confirms that when the
// payload is valid but the mailer cannot reach the SMTP server, the
// handler returns a retryable error (i.e. NOT wrapped in SkipRetry).
func TestHandleEmailSend_ValidPayload_DialError(t *testing.T) {
	w := NewEmailWorker(newTestMailer(t))

	payload := &tasks.EmailPayload{
		FromEmail: "from@example.com",
		ToEmail:   []string{"to@example.com"},
		Subject:   "Test",
		HTMLBody:  "<p>hello</p>",
	}
	task, err := tasks.NewEmailTask(payload)
	if err != nil {
		t.Fatalf("tasks.NewEmailTask: %v", err)
	}

	handlerErr := w.HandleEmailSend(context.Background(), task)

	// A dial failure is a transient error — it must NOT be wrapped in
	// SkipRetry so that asynq can schedule a retry.
	if handlerErr == nil {
		t.Fatal("expected a dial error, got nil")
	}
	if errors.Is(handlerErr, asynq.SkipRetry) {
		t.Errorf("dial error should be retryable but got SkipRetry: %v", handlerErr)
	}
}

// TestHandleEmailSend_CancelledContext verifies that a cancelled context
// propagates as an error (not SkipRetry), letting the caller decide
// whether to retry.
func TestHandleEmailSend_CancelledContext(t *testing.T) {
	w := NewEmailWorker(newTestMailer(t))

	payload := &tasks.EmailPayload{
		FromEmail: "from@example.com",
		ToEmail:   []string{"to@example.com"},
		Subject:   "ctx cancel test",
		HTMLBody:  "<p>bye</p>",
	}
	task, err := tasks.NewEmailTask(payload)
	if err != nil {
		t.Fatalf("tasks.NewEmailTask: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	handlerErr := w.HandleEmailSend(ctx, task)

	if handlerErr == nil {
		t.Fatal("expected an error for cancelled context, got nil")
	}
	if errors.Is(handlerErr, asynq.SkipRetry) {
		t.Errorf("context-cancelled error should be retryable, got SkipRetry: %v", handlerErr)
	}
}
