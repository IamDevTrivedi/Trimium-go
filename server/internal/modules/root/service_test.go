package root

import (
	"context"
	"errors"
	"testing"

	"server/internal/tasks"
)

type mockRepository struct{}

type fakeQueue struct {
	payloads []tasks.EmailPayload
	err      error
}

func (f *fakeQueue) SendEmail(_ context.Context, payload tasks.EmailPayload) error {
	f.payloads = append(f.payloads, payload)
	return f.err
}

func TestNewService(t *testing.T) {
	if svc := NewService(&mockRepository{}, nil); svc == nil {
		t.Fatal("expected NewService to return a non-nil service")
	}
}

func TestService_GetAppInfo(t *testing.T) {
	svc := NewService(&mockRepository{}, nil)

	got, err := svc.GetAppInfo(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil app info")
	}

	want := &AppInfo{
		Name:    "Trimium API",
		Version: "1.0.0",
		Status:  "ok",
	}

	if got.Name != want.Name {
		t.Errorf("expected Name %q, got %q", want.Name, got.Name)
	}
	if got.Version != want.Version {
		t.Errorf("expected Version %q, got %q", want.Version, got.Version)
	}
	if got.Status != want.Status {
		t.Errorf("expected Status %q, got %q", want.Status, got.Status)
	}
}

func TestService_SendWelcomeEmail_BuildsPayload(t *testing.T) {
	fq := &fakeQueue{}
	svc := NewService(&mockRepository{}, fq)

	to := []string{"user1@example.com", "user2@example.com"}
	if err := svc.SendWelcomeEmail(context.Background(), to); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(fq.payloads) != 1 {
		t.Fatalf("expected 1 dispatched payload, got %d", len(fq.payloads))
	}

	got := fq.payloads[0]
	if got.FromEmail != welcomeFromEmail {
		t.Errorf("expected FromEmail %q, got %q", welcomeFromEmail, got.FromEmail)
	}
	if !equalStrings(got.ToEmail, to) {
		t.Errorf("expected ToEmail %v, got %v", to, got.ToEmail)
	}
	if got.Subject != welcomeSubject {
		t.Errorf("expected Subject %q, got %q", welcomeSubject, got.Subject)
	}
	if got.HTMLBody != welcomeHTMLBody {
		t.Errorf("expected HTMLBody %q, got %q", welcomeHTMLBody, got.HTMLBody)
	}
}

func TestService_SendWelcomeEmail_PropagatesQueueError(t *testing.T) {
	sentinel := errors.New("queue down")
	svc := NewService(&mockRepository{}, &fakeQueue{err: sentinel})

	if err := svc.SendWelcomeEmail(context.Background(), []string{"user@example.com"}); !errors.Is(err, sentinel) {
		t.Fatalf("expected error %v, got %v", sentinel, err)
	}
}

func TestService_SendWelcomeEmail_RejectsEmptyRecipients(t *testing.T) {
	fq := &fakeQueue{}
	svc := NewService(&mockRepository{}, fq)

	if err := svc.SendWelcomeEmail(context.Background(), nil); err == nil {
		t.Fatal("expected an error for empty recipients")
	}
	if len(fq.payloads) != 0 {
		t.Errorf("expected no dispatch, got %d payloads", len(fq.payloads))
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
