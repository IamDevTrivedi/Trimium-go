package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// unreachableMongoOptions returns options that pass validation but point at
// a port where nothing listens, so tests never need a running MongoDB.
func unreachableMongoOptions() MongoOptions {
	return MongoOptions{
		URI:            "mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=100&connectTimeoutMS=100",
		Database:       "testdb",
		MaxPoolSize:    10,
		MinPoolSize:    0,
		ConnectTimeout: 100 * time.Millisecond,
		PingTimeout:    250 * time.Millisecond,
	}
}

func TestNewMongo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*MongoOptions)
		wantErr string
	}{
		{
			name:    "empty uri",
			mutate:  func(o *MongoOptions) { o.URI = "" },
			wantErr: "mongo: uri is required",
		},
		{
			name:    "empty database",
			mutate:  func(o *MongoOptions) { o.Database = "" },
			wantErr: "mongo: database name is required",
		},
		{
			name: "min pool size greater than max pool size",
			mutate: func(o *MongoOptions) {
				o.MaxPoolSize = 10
				o.MinPoolSize = 20
			},
			wantErr: "mongo: min pool size 20 cannot exceed max pool size 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := unreachableMongoOptions()
			tt.mutate(&opts)

			m, err := NewMongo(context.Background(), opts)
			if err == nil {
				_ = m.Close(context.Background())
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
			if m != nil {
				t.Fatalf("expected nil *Mongo on error, got %+v", m)
			}
		})
	}
}

func TestNewMongo_InvalidURI(t *testing.T) {
	opts := unreachableMongoOptions()
	opts.URI = "mongodb://" // no hosts: rejected by the driver's URI parser

	m, err := NewMongo(context.Background(), opts)
	if err == nil {
		_ = m.Close(context.Background())
		t.Fatal("expected connect error, got nil")
	}
	if !strings.Contains(err.Error(), "mongo: connect:") {
		t.Fatalf("expected error containing %q, got %q", "mongo: connect:", err.Error())
	}
}

func TestNewMongo_PingFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode: dials an unreachable local port")
	}

	opts := unreachableMongoOptions()

	m, err := NewMongo(context.Background(), opts)
	if err == nil {
		_ = m.Close(context.Background())
		t.Fatal("expected ping error, got nil")
	}
	if !strings.Contains(err.Error(), "mongo: ping:") {
		t.Fatalf("expected error containing %q, got %q", "mongo: ping:", err.Error())
	}
}

func TestMongo_Accessors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode: dials an unreachable local port")
	}

	// Connect is lazy: building a client never touches the network, so this
	// works without a running MongoDB.
	client, err := mongo.Connect(options.Client().
		ApplyURI("mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=100&connectTimeoutMS=100"))
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}

	m := &Mongo{
		client: client,
		db:     client.Database("testdb"),
	}

	if m.Client() == nil {
		t.Error("Client() returned nil")
	}
	if m.DB() == nil {
		t.Error("DB() returned nil")
	}
	if m.Collection("users") == nil {
		t.Error(`Collection("users") returned nil`)
	}

	// Nothing listens on port 1, so the ping must fail quickly.
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := m.Ping(ctx); err == nil {
		t.Error("expected ping against unreachable server to fail, got nil")
	}

	if err := m.Close(context.Background()); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
