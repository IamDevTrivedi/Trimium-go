package health

import (
	"testing"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected NewService() to return non-nil service instance")
	}
}

func TestService_GetStatus(t *testing.T) {
	svc := NewService()

	data, err := svc.GetStatus()
	if err != nil {
		t.Fatalf("expected GetStatus() to return nil error, got %v", err)
	}
	if data == nil {
		t.Fatal("expected GetStatus() to return non-nil payload")
	}

	if got := data["status"]; got != "ok" {
		t.Errorf("expected status %q, got %v", "ok", got)
	}
}
