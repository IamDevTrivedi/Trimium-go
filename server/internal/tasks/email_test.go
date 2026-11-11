package tasks

import (
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
)

// helpers

func mustNewEmailTask(t *testing.T, p *EmailPayload) *asynq.Task {
	t.Helper()
	task, err := NewEmailTask(p)
	if err != nil {
		t.Fatalf("NewEmailTask: unexpected error: %v", err)
	}
	return task
}

// NewEmailTask

func TestNewEmailTask_SetsCorrectType(t *testing.T) {
	task := mustNewEmailTask(t, &EmailPayload{
		FromEmail: "from@example.com",
		ToEmail:   []string{"to@example.com"},
		Subject:   "Hello",
		HTMLBody:  "<p>Hi</p>",
	})

	if task.Type() != TypeEmailSend {
		t.Errorf("expected type %q, got %q", TypeEmailSend, task.Type())
	}
}

func TestNewEmailTask_PayloadIsValidJSON(t *testing.T) {
	p := &EmailPayload{
		FromEmail: "from@example.com",
		ToEmail:   []string{"a@example.com", "b@example.com"},
		Subject:   "Test Subject",
		HTMLBody:  "<h1>Test</h1>",
	}

	task := mustNewEmailTask(t, p)

	var decoded EmailPayload
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
}

func TestNewEmailTask_PayloadRoundTrips(t *testing.T) {
	want := &EmailPayload{
		FromEmail: "sender@example.com",
		ToEmail:   []string{"r1@example.com", "r2@example.com"},
		Subject:   "Round-trip subject",
		HTMLBody:  "<p>body</p>",
	}

	task := mustNewEmailTask(t, want)

	var got EmailPayload
	if err := json.Unmarshal(task.Payload(), &got); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if got.FromEmail != want.FromEmail {
		t.Errorf("FromEmail: want %q, got %q", want.FromEmail, got.FromEmail)
	}
	if got.Subject != want.Subject {
		t.Errorf("Subject: want %q, got %q", want.Subject, got.Subject)
	}
	if got.HTMLBody != want.HTMLBody {
		t.Errorf("HTMLBody: want %q, got %q", want.HTMLBody, got.HTMLBody)
	}
	if len(got.ToEmail) != len(want.ToEmail) {
		t.Fatalf("ToEmail length: want %d, got %d", len(want.ToEmail), len(got.ToEmail))
	}
	for i := range want.ToEmail {
		if got.ToEmail[i] != want.ToEmail[i] {
			t.Errorf("ToEmail[%d]: want %q, got %q", i, want.ToEmail[i], got.ToEmail[i])
		}
	}
}

func TestNewEmailTask_ZeroValuePayload(t *testing.T) {
	task := mustNewEmailTask(t, &EmailPayload{})

	if task == nil {
		t.Fatal("expected non-nil task for zero-value payload")
	}
	if task.Type() != TypeEmailSend {
		t.Errorf("expected type %q, got %q", TypeEmailSend, task.Type())
	}
}

// ParseEmailPayload

func TestParseEmailPayload_RoundTrip(t *testing.T) {
	want := &EmailPayload{
		FromEmail: "from@example.com",
		ToEmail:   []string{"to@example.com"},
		Subject:   "Parse test",
		HTMLBody:  "<p>parse</p>",
	}

	task := mustNewEmailTask(t, want)

	got, err := ParseEmailPayload(task)
	if err != nil {
		t.Fatalf("ParseEmailPayload: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil payload")
	}

	if got.FromEmail != want.FromEmail {
		t.Errorf("FromEmail: want %q, got %q", want.FromEmail, got.FromEmail)
	}
	if got.Subject != want.Subject {
		t.Errorf("Subject: want %q, got %q", want.Subject, got.Subject)
	}
	if got.HTMLBody != want.HTMLBody {
		t.Errorf("HTMLBody: want %q, got %q", want.HTMLBody, got.HTMLBody)
	}
	if len(got.ToEmail) != len(want.ToEmail) {
		t.Fatalf("ToEmail length: want %d, got %d", len(want.ToEmail), len(got.ToEmail))
	}
	for i := range want.ToEmail {
		if got.ToEmail[i] != want.ToEmail[i] {
			t.Errorf("ToEmail[%d]: want %q, got %q", i, want.ToEmail[i], got.ToEmail[i])
		}
	}
}

func TestParseEmailPayload_InvalidJSON(t *testing.T) {
	task := asynq.NewTask(TypeEmailSend, []byte("not-json{{{"))

	got, err := ParseEmailPayload(task)
	if err == nil {
		t.Fatal("expected an error for invalid JSON payload, got nil")
	}
	if got != nil {
		t.Errorf("expected nil payload on error, got %+v", got)
	}
}

func TestParseEmailPayload_EmptyPayload(t *testing.T) {
	task := asynq.NewTask(TypeEmailSend, []byte(""))

	got, err := ParseEmailPayload(task)
	if err == nil {
		t.Fatal("expected an error for empty payload, got nil")
	}
	if got != nil {
		t.Errorf("expected nil payload on error, got %+v", got)
	}
}

func TestParseEmailPayload_MultipleRecipients(t *testing.T) {
	want := &EmailPayload{
		FromEmail: "bulk@example.com",
		ToEmail:   []string{"a@x.com", "b@x.com", "c@x.com"},
		Subject:   "Bulk",
		HTMLBody:  "<p>bulk</p>",
	}

	task := mustNewEmailTask(t, want)
	got, err := ParseEmailPayload(task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.ToEmail) != len(want.ToEmail) {
		t.Fatalf("expected %d recipients, got %d", len(want.ToEmail), len(got.ToEmail))
	}
}

// TypeEmailSend constant

func TestTypeEmailSend_Value(t *testing.T) {
	if TypeEmailSend == "" {
		t.Error("TypeEmailSend must not be empty")
	}
	// Ensure the constant is stable so queue consumers can rely on it.
	const expected = "email:send"
	if TypeEmailSend != expected {
		t.Errorf("TypeEmailSend: want %q, got %q", expected, TypeEmailSend)
	}
}
