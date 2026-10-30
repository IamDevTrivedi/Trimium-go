package root

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

func (noopHandler) HandleRoot(c *gin.Context) { c.Status(http.StatusNoContent) }

func serveRootRequest(t *testing.T, router *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1", nil)
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

func TestRateLimiter_AllowsFiftyRequestsPerMinute(t *testing.T) {
	store, mr := newTestRedis(t)

	const ip = "203.0.113.50"

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestIP())
	RegisterRoutes(router, NewRateLimiter(store), noopHandler{})

	for i := range rateLimitPerMinute {
		w := serveRootRequest(t, router, ip+":1234")
		if w.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected status %d, got %d: %s",
				i+1, http.StatusNoContent, w.Code, w.Body.String())
		}
	}

	w := serveRootRequest(t, router, ip+":1234")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request %d: expected status %d, got %d: %s",
			rateLimitPerMinute+1, http.StatusTooManyRequests, w.Code, w.Body.String())
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

	key := "rl:root:index:" + ip
	if !mr.Exists(key) {
		t.Errorf("expected rate-limit key %q in redis", key)
	}
}

func TestRateLimiter_IsolatesClients(t *testing.T) {
	store, _ := newTestRedis(t)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestIP())
	RegisterRoutes(router, NewRateLimiter(store), noopHandler{})

	for range rateLimitPerMinute {
		if w := serveRootRequest(t, router, "203.0.113.60:1234"); w.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	}
	if w := serveRootRequest(t, router, "203.0.113.60:1234"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d: %s", http.StatusTooManyRequests, w.Code, w.Body.String())
	}
	if w := serveRootRequest(t, router, "203.0.113.61:1234"); w.Code != http.StatusNoContent {
		t.Fatalf("expected a different client to be allowed, got %d: %s", w.Code, w.Body.String())
	}
}
