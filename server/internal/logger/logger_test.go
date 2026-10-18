package logger_test

import (
	"testing"

	"server/internal/logger"
)

func TestGet(t *testing.T) {
	log1 := logger.Get()
	if log1 == nil {
		t.Fatal("expected logger instance, got nil")
	}

	log2 := logger.Get()
	if log2 != log1 {
		t.Errorf("expected Get() to return the same instance, got %p and %p", log1, log2)
	}
}
