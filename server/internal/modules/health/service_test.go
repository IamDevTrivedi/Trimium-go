package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	pingRedis func(context.Context) error
	pingMongo func(context.Context) error
}

func (m *mockRepository) PingRedis(ctx context.Context) error {
	if m.pingRedis == nil {
		return nil
	}
	return m.pingRedis(ctx)
}

func (m *mockRepository) PingMongo(ctx context.Context) error {
	if m.pingMongo == nil {
		return nil
	}
	return m.pingMongo(ctx)
}

func TestNewService(t *testing.T) {
	if svc := NewService(&mockRepository{}); svc == nil {
		t.Fatal("expected NewService to return a non-nil service")
	}
}

func TestService_GetAppStatus(t *testing.T) {
	called := false
	svc := NewService(&mockRepository{
		pingRedis: func(context.Context) error { called = true; return nil },
		pingMongo: func(context.Context) error { called = true; return nil },
	})

	got, err := svc.GetAppStatus(t.Context())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got == nil || got.Status != "ok" {
		t.Fatalf("expected status %q, got %#v", "ok", got)
	}
	if called {
		t.Error("expected app status not to touch the repository")
	}
}

func TestService_DatastoreStatus(t *testing.T) {
	sentinel := errors.New("datastore down")

	tests := []struct {
		name   string
		get    func(Service) (*StatusResponse, error)
		ping   func(*mockRepository, func(context.Context) error)
		wantOK bool
	}{
		{
			name:   "redis healthy",
			get:    func(s Service) (*StatusResponse, error) { return s.GetRedisStatus(t.Context()) },
			ping:   func(m *mockRepository, fn func(context.Context) error) { m.pingRedis = fn },
			wantOK: true,
		},
		{
			name:   "redis unhealthy",
			get:    func(s Service) (*StatusResponse, error) { return s.GetRedisStatus(t.Context()) },
			ping:   func(m *mockRepository, fn func(context.Context) error) { m.pingRedis = fn },
			wantOK: false,
		},
		{
			name:   "mongo healthy",
			get:    func(s Service) (*StatusResponse, error) { return s.GetMongoStatus(t.Context()) },
			ping:   func(m *mockRepository, fn func(context.Context) error) { m.pingMongo = fn },
			wantOK: true,
		},
		{
			name:   "mongo unhealthy",
			get:    func(s Service) (*StatusResponse, error) { return s.GetMongoStatus(t.Context()) },
			ping:   func(m *mockRepository, fn func(context.Context) error) { m.pingMongo = fn },
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			if tt.wantOK {
				tt.ping(repo, func(context.Context) error { return nil })
			} else {
				tt.ping(repo, func(context.Context) error { return sentinel })
			}

			got, err := tt.get(NewService(repo))

			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil || got.Status != "ok" {
					t.Fatalf("expected status %q, got %#v", "ok", got)
				}
				return
			}

			if !errors.Is(err, sentinel) {
				t.Fatalf("expected error %v, got %v", sentinel, err)
			}
			if got != nil {
				t.Errorf("expected nil status on error, got %#v", got)
			}
		})
	}
}

func TestService_DatastorePingUsesTimeout(t *testing.T) {
	tests := []struct {
		name string
		get  func(Service, context.Context) (*StatusResponse, error)
		ping func(*mockRepository, func(context.Context) error)
	}{
		{
			name: "redis",
			get:  func(s Service, ctx context.Context) (*StatusResponse, error) { return s.GetRedisStatus(ctx) },
			ping: func(m *mockRepository, fn func(context.Context) error) { m.pingRedis = fn },
		},
		{
			name: "mongo",
			get:  func(s Service, ctx context.Context) (*StatusResponse, error) { return s.GetMongoStatus(ctx) },
			ping: func(m *mockRepository, fn func(context.Context) error) { m.pingMongo = fn },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var observed time.Duration

			repo := &mockRepository{}
			tt.ping(repo, func(ctx context.Context) error {
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatal("expected ping context to carry a deadline")
				}
				observed = time.Until(deadline)
				return nil
			})

			if _, err := tt.get(NewService(repo), t.Context()); err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if observed <= 0 || observed > pingTimeout {
				t.Errorf("expected deadline within (0, %v], got %v", pingTimeout, observed)
			}
		})
	}
}
