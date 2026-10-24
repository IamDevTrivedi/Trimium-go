package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions configures the Redis client connection pool and timeouts.
type RedisOptions struct {
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingTimeout  time.Duration
}

// Redis wraps a Redis client.
type Redis struct {
	client *redis.Client
}

// NewRedis creates a Redis client, verifies the connection with a ping and
// returns it.
func NewRedis(ctx context.Context, addr, username, password string, db int, opts RedisOptions) (*Redis, error) {
	if addr == "" {
		return nil, errors.New("redis: addr is required")
	}
	if db < 0 {
		return nil, fmt.Errorf("redis: db %d cannot be negative", db)
	}
	if opts.PoolSize <= 0 {
		return nil, fmt.Errorf("redis: pool size %d must be greater than zero", opts.PoolSize)
	}
	if opts.MinIdleConns < 0 {
		return nil, fmt.Errorf("redis: min idle conns %d cannot be negative", opts.MinIdleConns)
	}
	if opts.MinIdleConns > opts.PoolSize {
		return nil, fmt.Errorf(
			"redis: min idle conns %d cannot exceed pool size %d",
			opts.MinIdleConns,
			opts.PoolSize,
		)
	}
	if opts.DialTimeout <= 0 {
		return nil, fmt.Errorf("redis: dial timeout %v must be greater than zero", opts.DialTimeout)
	}
	if opts.ReadTimeout <= 0 {
		return nil, fmt.Errorf("redis: read timeout %v must be greater than zero", opts.ReadTimeout)
	}
	if opts.WriteTimeout <= 0 {
		return nil, fmt.Errorf("redis: write timeout %v must be greater than zero", opts.WriteTimeout)
	}
	if opts.PingTimeout <= 0 {
		return nil, fmt.Errorf("redis: ping timeout %v must be greater than zero", opts.PingTimeout)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Username:     username,
		Password:     password,
		DB:           db,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})

	r := &Redis{client: client}

	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()

	if err := r.Ping(pingCtx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return r, nil
}

// Client returns the underlying Redis client.
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Ping checks that the deployment is reachable.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the client and releases its connections.
func (r *Redis) Close() error {
	return r.client.Close()
}
