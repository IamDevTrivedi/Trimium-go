package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// unreachableRedisOptions returns options that pass validation but point at
// a port where nothing listens, so tests never need a running Redis.
func unreachableRedisOptions() RedisOptions {
	return RedisOptions{
		Host:         "127.0.0.1",
		Port:         1,
		PoolSize:     10,
		MinIdleConns: 0,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PingTimeout:  500 * time.Millisecond,
	}
}

func TestNewRedis_Validation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*RedisOptions)
		wantErr string
	}{
		{
			name:    "empty host",
			mutate:  func(o *RedisOptions) { o.Host = "" },
			wantErr: "redis: host is required",
		},
		{
			name:    "invalid port zero",
			mutate:  func(o *RedisOptions) { o.Port = 0 },
			wantErr: "redis: invalid port 0",
		},
		{
			name:    "invalid port too high",
			mutate:  func(o *RedisOptions) { o.Port = 99999 },
			wantErr: "redis: invalid port 99999",
		},
		{
			name:    "pool size zero",
			mutate:  func(o *RedisOptions) { o.PoolSize = 0 },
			wantErr: "redis: pool size 0 must be greater than zero",
		},
		{
			name:    "negative min idle conns",
			mutate:  func(o *RedisOptions) { o.MinIdleConns = -1 },
			wantErr: "redis: min idle conns -1 cannot be negative",
		},
		{
			name: "min idle conns greater than pool size",
			mutate: func(o *RedisOptions) {
				o.PoolSize = 10
				o.MinIdleConns = 20
			},
			wantErr: "redis: min idle conns 20 cannot exceed pool size 10",
		},
		{
			name:    "dial timeout not set",
			mutate:  func(o *RedisOptions) { o.DialTimeout = 0 },
			wantErr: "redis: dial timeout 0s must be greater than zero",
		},
		{
			name:    "read timeout not set",
			mutate:  func(o *RedisOptions) { o.ReadTimeout = 0 },
			wantErr: "redis: read timeout 0s must be greater than zero",
		},
		{
			name:    "write timeout not set",
			mutate:  func(o *RedisOptions) { o.WriteTimeout = 0 },
			wantErr: "redis: write timeout 0s must be greater than zero",
		},
		{
			name:    "ping timeout not set",
			mutate:  func(o *RedisOptions) { o.PingTimeout = 0 },
			wantErr: "redis: ping timeout 0s must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := unreachableRedisOptions()
			tt.mutate(&opts)

			r, err := NewRedis(context.Background(), opts)
			if err == nil {
				_ = r.Close()
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
			if r != nil {
				t.Fatalf("expected nil *Redis on error, got %+v", r)
			}
		})
	}
}

func TestNewRedis_PingFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode: dials an unreachable local port")
	}

	opts := unreachableRedisOptions()

	r, err := NewRedis(context.Background(), opts)
	if err == nil {
		_ = r.Close()
		t.Fatal("expected ping error, got nil")
	}
	if !strings.Contains(err.Error(), "redis: ping:") {
		t.Fatalf("expected error containing %q, got %q", "redis: ping:", err.Error())
	}
}

func TestRedis_Accessors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode: dials an unreachable local port")
	}

	// NewClient is lazy: no connection is made until the first command.
	client := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 250 * time.Millisecond,
		ReadTimeout: 250 * time.Millisecond,
	})
	r := &Redis{client: client}

	if r.Client() != client {
		t.Error("Client() did not return the underlying client")
	}

	// Nothing listens on port 1, so the ping must fail quickly.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := r.Ping(ctx); err == nil {
		t.Error("expected ping against unreachable server to fail, got nil")
	}

	if err := r.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
