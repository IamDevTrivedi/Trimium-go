package root

import (
	"context"
	"testing"
)

type mockRepository struct{}

func TestNewService(t *testing.T) {
	if svc := NewService(&mockRepository{}); svc == nil {
		t.Fatal("expected NewService to return a non-nil service")
	}
}

func TestService_GetAppInfo(t *testing.T) {
	svc := NewService(&mockRepository{})

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
