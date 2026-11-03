package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOptions struct {
	URI          string
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingTimeout  time.Duration
}

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, opts RedisOptions) (*Redis, error) {
	if opts.URI == "" {
		return nil, errors.New("redis: uri is required")
	}
	redisOpts, err := redis.ParseURL(opts.URI)
	if err != nil {
		return nil, fmt.Errorf("redis: invalid uri: %w", err)
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

	redisOpts.PoolSize = opts.PoolSize
	redisOpts.MinIdleConns = opts.MinIdleConns
	redisOpts.DialTimeout = opts.DialTimeout
	redisOpts.ReadTimeout = opts.ReadTimeout
	redisOpts.WriteTimeout = opts.WriteTimeout

	client := redis.NewClient(redisOpts)

	r := &Redis{client: client}

	pingCtx, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()

	if err := r.Ping(pingCtx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return r, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}
