package health

import (
	"context"
	"errors"
	"testing"
)

type stubPinger struct {
	err  error
	ctx  context.Context
	runs int
}

func (p *stubPinger) Ping(ctx context.Context) error {
	p.runs++
	p.ctx = ctx
	return p.err
}

func TestNewRepository(t *testing.T) {
	if repo := NewRepository(&stubPinger{}, &stubPinger{}); repo == nil {
		t.Fatal("expected NewRepository to return a non-nil repository")
	}
}

func TestRepository_PingRedis(t *testing.T) {
	sentinel := errors.New("redis down")

	tests := []struct {
		name    string
		pingErr error
		wantErr error
	}{
		{name: "healthy", pingErr: nil, wantErr: nil},
		{name: "unhealthy", pingErr: sentinel, wantErr: sentinel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := &stubPinger{err: tt.pingErr}
			mongo := &stubPinger{}
			repo := NewRepository(redis, mongo)

			err := repo.PingRedis(t.Context())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if redis.runs != 1 {
				t.Errorf("expected redis pinger to run once, ran %d times", redis.runs)
			}
			if mongo.runs != 0 {
				t.Errorf("expected mongo pinger not to run, ran %d times", mongo.runs)
			}
			if redis.ctx == nil {
				t.Error("expected context to be forwarded to the redis pinger")
			}
		})
	}
}

func TestRepository_PingMongo(t *testing.T) {
	sentinel := errors.New("mongo down")

	tests := []struct {
		name    string
		pingErr error
		wantErr error
	}{
		{name: "healthy", pingErr: nil, wantErr: nil},
		{name: "unhealthy", pingErr: sentinel, wantErr: sentinel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := &stubPinger{}
			mongo := &stubPinger{err: tt.pingErr}
			repo := NewRepository(redis, mongo)

			err := repo.PingMongo(t.Context())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if mongo.runs != 1 {
				t.Errorf("expected mongo pinger to run once, ran %d times", mongo.runs)
			}
			if redis.runs != 0 {
				t.Errorf("expected redis pinger not to run, ran %d times", redis.runs)
			}
			if mongo.ctx == nil {
				t.Error("expected context to be forwarded to the mongo pinger")
			}
		})
	}
}
