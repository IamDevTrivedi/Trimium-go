package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"server/internal/database"
	"server/internal/middleware"
	"server/internal/response"
)

func newTestRedis(t *testing.T) (*database.Redis, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	store, err := database.NewRedis(context.Background(), "redis://"+mr.Addr(), database.RedisOptions{
		PoolSize:     5,
		MinIdleConns: 1,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PingTimeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create redis test store: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil && !errors.Is(err, redis.ErrClosed) {
			t.Errorf("close redis store: %v", err)
		}
	})

	return store, mr
}

type noopHandler struct{}

func (noopHandler) HandleHealth(c *gin.Context)      { c.Status(http.StatusNoContent) }
func (noopHandler) HandleRedisHealth(c *gin.Context) { c.Status(http.StatusNoContent) }
func (noopHandler) HandleMongoHealth(c *gin.Context) { c.Status(http.StatusNoContent) }

func serveHealthRequest(t *testing.T, router *gin.Engine, path, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	req.RemoteAddr = remoteAddr

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestNewRateLimiter(t *testing.T) {
	store, _ := newTestRedis(t)

	if rl := NewRateLimiter(store); rl == nil {
		t.Fatal("expected NewRateLimiter to return a non-nil rate limiter")
	}
}

func TestRateLimiter_EnforcesPerEndpointLimit(t *testing.T) {
	store, _ := newTestRedis(t)

	tests := []struct {
		name  string
		path  string
		ip    string
		limit int
	}{
		{name: "health", path: "/api/v1/healthz", ip: "203.0.113.10", limit: 5},
		{name: "redis", path: "/api/v1/healthz/redis", ip: "203.0.113.11", limit: 10},
		{name: "mongo", path: "/api/v1/healthz/mongo", ip: "203.0.113.12", limit: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.RequestID(), middleware.RequestIP())
			RegisterRoutes(router, NewRateLimiter(store), noopHandler{})

			for i := range tt.limit {
				w := serveHealthRequest(t, router, tt.path, tt.ip+":1234")
				if w.Code != http.StatusNoContent {
					t.Fatalf("request %d: expected status %d, got %d: %s",
						i+1, http.StatusNoContent, w.Code, w.Body.String())
				}
			}

			w := serveHealthRequest(t, router, tt.path, tt.ip+":1234")
			if w.Code != http.StatusTooManyRequests {
				t.Fatalf("request %d: expected status %d, got %d: %s",
					tt.limit+1, http.StatusTooManyRequests, w.Code, w.Body.String())
			}

			var got response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}
			if got.Success {
				t.Error("expected success to be false")
			}
			if got.Msg != "Rate limit exceeded" {
				t.Errorf("expected msg %q, got %q", "Rate limit exceeded", got.Msg)
			}
		})
	}
}

func TestRateLimiter_IsolatesClients(t *testing.T) {
	store, _ := newTestRedis(t)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestIP())
	RegisterRoutes(router, NewRateLimiter(store), noopHandler{})

	const path = "/api/v1/healthz"

	for range 5 {
		if w := serveHealthRequest(t, router, path, "203.0.113.20:1234"); w.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	}
	if w := serveHealthRequest(t, router, path, "203.0.113.20:1234"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d: %s", http.StatusTooManyRequests, w.Code, w.Body.String())
	}
	if w := serveHealthRequest(t, router, path, "203.0.113.21:1234"); w.Code != http.StatusNoContent {
		t.Fatalf("expected a different client to be allowed, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRateLimiter_UsesDedicatedKeyPrefixes(t *testing.T) {
	store, mr := newTestRedis(t)

	const ip = "203.0.113.30"

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestIP())
	RegisterRoutes(router, NewRateLimiter(store), noopHandler{})

	for _, path := range []string{"/api/v1/healthz", "/api/v1/healthz/redis", "/api/v1/healthz/mongo"} {
		if w := serveHealthRequest(t, router, path, ip+":1234"); w.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	}

	want := []string{
		"rl:health:index:" + ip,
		"rl:health:redis:" + ip,
		"rl:health:mongo:" + ip,
	}
	got := mr.Keys()

	if len(got) != len(want) {
		t.Fatalf("expected %d keys %v, got %d %v", len(want), want, len(got), got)
	}

	seen := make(map[string]struct{}, len(got))
	for _, key := range got {
		seen[key] = struct{}{}
	}
	for _, key := range want {
		if _, ok := seen[key]; !ok {
			t.Errorf("expected key %q in redis, got %v", key, got)
		}
	}
}
