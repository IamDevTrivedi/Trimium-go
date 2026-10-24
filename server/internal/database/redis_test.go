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

func TestNewRedisRequiresAddr(t *testing.T) {
	_, err := NewRedis(context.Background(), "", "", "", 0, validRedisOptions())
	if err == nil {
		t.Fatal("expected error for empty addr, got nil")
	}
	if !strings.Contains(err.Error(), "addr is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewRedisRejectsNegativeDB(t *testing.T) {
	_, err := NewRedis(context.Background(), "localhost:6379", "", "", -1, validRedisOptions())
	if err == nil {
		t.Fatal("expected error for negative db, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be negative") {
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

			_, err := NewRedis(context.Background(), "localhost:6379", "", "", 0, opts)
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

			_, err := NewRedis(context.Background(), "localhost:6379", "", "", 0, opts)
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

			_, err := NewRedis(context.Background(), "localhost:6379", "", "", 0, opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
