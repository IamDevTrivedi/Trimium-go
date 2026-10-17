package root

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
	status := svc.GetStatus()

	if status == nil {
		t.Fatal("expected GetStatus() to return non-nil map")
	}

	expectedMessage := "Welcome to Trimium API"
	if msg, ok := status["message"]; !ok || msg != expectedMessage {
		t.Errorf("expected message %q, got %v", expectedMessage, msg)
	}

	expectedStatus := "ok"
	if st, ok := status["status"]; !ok || st != expectedStatus {
		t.Errorf("expected status %q, got %v", expectedStatus, st)
	}
}
