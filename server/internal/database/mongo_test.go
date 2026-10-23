package database

import (
	"context"
	"strings"
	"testing"
)

func TestNewMongoRequiresURI(t *testing.T) {
	_, err := NewMongo(context.Background(), "", "trimium", MongoOptions{})
	if err == nil {
		t.Fatal("expected error for empty uri, got nil")
	}
	if !strings.Contains(err.Error(), "uri is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewMongoRequiresDatabase(t *testing.T) {
	_, err := NewMongo(context.Background(), "mongodb://localhost:27017", "", MongoOptions{})
	if err == nil {
		t.Fatal("expected error for empty database, got nil")
	}
	if !strings.Contains(err.Error(), "database name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewMongoRejectsMinPoolSizeAboveMax(t *testing.T) {
	_, err := NewMongo(context.Background(), "mongodb://localhost:27017", "trimium", MongoOptions{
		MaxPoolSize: 5,
		MinPoolSize: 10,
	})
	if err == nil {
		t.Fatal("expected error for min pool size above max, got nil")
	}
	if !strings.Contains(err.Error(), "cannot exceed max pool size") {
		t.Errorf("unexpected error: %v", err)
	}
}
