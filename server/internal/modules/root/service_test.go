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

	data, err := svc.GetStatus()
	if err != nil {
		t.Fatalf("expected GetStatus() to return nil error, got %v", err)
	}

	// The service is currently a placeholder that returns no payload;
	// the welcome message is owned by the handler. Update this test
	// once GetStatus starts returning real status data.
	if data != nil {
		t.Errorf("expected GetStatus() to return nil payload, got %v", data)
	}

	// Calling twice must not introduce state or fail.
	if _, err := svc.GetStatus(); err != nil {
		t.Errorf("expected second GetStatus() call to return nil error, got %v", err)
	}
}
