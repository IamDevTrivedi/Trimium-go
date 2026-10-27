package database

import (
	"context"
	"strings"
	"testing"
	"time"
)

func validRedisOptions() RedisOptions {
	return RedisOptions{
		PoolSize:     10,
		MinIdleConns: 1,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PingTimeout:  time.Second,
	}
}

func TestNewRedisRequiresURI(t *testing.T) {
	_, err := NewRedis(context.Background(), "", validRedisOptions())
	if err == nil {
		t.Fatal("expected error for empty uri, got nil")
	}
	if !strings.Contains(err.Error(), "uri is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRedisRejectsInvalidURI(t *testing.T) {
	_, err := NewRedis(context.Background(), "invalid-uri", validRedisOptions())
	if err == nil {
		t.Fatal("expected error for invalid uri, got nil")
	}
	if !strings.Contains(err.Error(), "invalid uri") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRedisRejectsInvalidPoolSize(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validRedisOptions()
			opts.PoolSize = tt.poolSize

			_, err := NewRedis(context.Background(), "redis://localhost:6379", opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "pool size") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewRedisRejectsInvalidMinIdleConns(t *testing.T) {
	tests := []struct {
		name         string
		minIdleConns int
		want         string
	}{
		{"negative", -1, "cannot be negative"},
		{"exceeds pool size", 11, "cannot exceed pool size"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validRedisOptions()
			opts.MinIdleConns = tt.minIdleConns

			_, err := NewRedis(context.Background(), "redis://localhost:6379", opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewRedisRejectsInvalidTimeouts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*RedisOptions)
		want   string
	}{
		{"zero dial timeout", func(o *RedisOptions) { o.DialTimeout = 0 }, "dial timeout"},
		{"negative dial timeout", func(o *RedisOptions) { o.DialTimeout = -time.Second }, "dial timeout"},
		{"zero read timeout", func(o *RedisOptions) { o.ReadTimeout = 0 }, "read timeout"},
		{"negative read timeout", func(o *RedisOptions) { o.ReadTimeout = -time.Second }, "read timeout"},
		{"zero write timeout", func(o *RedisOptions) { o.WriteTimeout = 0 }, "write timeout"},
		{"negative write timeout", func(o *RedisOptions) { o.WriteTimeout = -time.Second }, "write timeout"},
		{"zero ping timeout", func(o *RedisOptions) { o.PingTimeout = 0 }, "ping timeout"},
		{"negative ping timeout", func(o *RedisOptions) { o.PingTimeout = -time.Second }, "ping timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validRedisOptions()
			tt.mutate(&opts)

			_, err := NewRedis(context.Background(), "redis://localhost:6379", opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
