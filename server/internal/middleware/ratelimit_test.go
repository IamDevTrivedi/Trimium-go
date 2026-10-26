package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"server/internal/response"
)

func newRateLimiterRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	t.Cleanup(func() {
		if err := client.Close(); err != nil && !errors.Is(err, redis.ErrClosed) {
			t.Errorf("close redis client: %v", err)
		}
	})

	return client, mr
}

func serveRateLimitRequest(t *testing.T, router *gin.Engine, path string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func requireRateLimitStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()

	if w.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, w.Code, w.Body.String())
	}
}

func TestRateLimiter_AllowsLimitThenReturns429(t *testing.T) {
	client, _ := newRateLimiterRedis(t)

	var handlerRuns int
	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  2,
			Window: time.Minute,
			KeyGen: func(*gin.Context) string {
				return "client-a"
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		handlerRuns++
		c.Status(http.StatusNoContent)
	})

	for range 2 {
		requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
	}
	if handlerRuns != 2 {
		t.Fatalf("expected handler to run 2 times, ran %d times", handlerRuns)
	}

	w := serveRateLimitRequest(t, router, "/test", nil)
	requireRateLimitStatus(t, w, http.StatusTooManyRequests)
	if handlerRuns != 2 {
		t.Errorf("expected handler to remain at 2 runs, ran %d times", handlerRuns)
	}

	var got response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Success {
		t.Error("expected success to be false")
	}
	if got.Msg != "Rate limit exceeded" {
		t.Errorf("expected message %q, got %q", "Rate limit exceeded", got.Msg)
	}
	if got.Metadata != nil {
		t.Errorf("expected nil metadata, got %#v", got.Metadata)
	}
}

func TestRateLimiter_UsesRequestIDAsLogMember(t *testing.T) {
	client, mr := newRateLimiterRedis(t)

	var requestID string
	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  1,
			Window: time.Minute,
			KeyGen: func(*gin.Context) string {
				return "client-a"
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		requestID = GetRequestID(c)
		c.Status(http.StatusNoContent)
	})

	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
	if requestID == "" {
		t.Fatal("expected RequestID middleware to set a request ID")
	}

	members, err := mr.ZMembers(rateLimitRedisKey("client-a"))
	if err != nil {
		t.Fatalf("read rate-limit log: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(members))
	}
	if members[0] != requestID {
		t.Errorf("expected log member %q, got %q", requestID, members[0])
	}
}

func TestRateLimiter_SlidingWindow(t *testing.T) {
	client, mr := newRateLimiterRedis(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	mr.SetTime(base)

	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  2,
			Window: time.Second,
			KeyGen: func(*gin.Context) string {
				return "client-a"
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)

	mr.SetTime(base.Add(900 * time.Millisecond))
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusTooManyRequests)

	mr.SetTime(base.Add(1_001 * time.Millisecond))
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusTooManyRequests)
}

func TestRateLimiter_IsolatesKeys(t *testing.T) {
	client, _ := newRateLimiterRedis(t)

	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  1,
			Window: time.Minute,
			KeyGen: func(c *gin.Context) string {
				return c.GetHeader("X-API-Key")
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	clientA := map[string]string{"X-API-Key": "client-a"}
	clientB := map[string]string{"X-API-Key": "client-b"}

	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", clientA), http.StatusNoContent)
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", clientA), http.StatusTooManyRequests)
	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", clientB), http.StatusNoContent)
}

func TestRateLimiter_RegistrationScopes(t *testing.T) {
	tests := []struct {
		name     string
		register func(router *gin.Engine, limiter, handler gin.HandlerFunc) string
	}{
		{
			name: "global",
			register: func(router *gin.Engine, limiter, handler gin.HandlerFunc) string {
				router.Use(limiter)
				router.GET("/scoped", handler)
				return "/scoped"
			},
		},
		{
			name: "group",
			register: func(router *gin.Engine, limiter, handler gin.HandlerFunc) string {
				group := router.Group("/api", limiter)
				group.GET("/scoped", handler)
				return "/api/scoped"
			},
		},
		{
			name: "route",
			register: func(router *gin.Engine, limiter, handler gin.HandlerFunc) string {
				router.GET("/scoped", limiter, handler)
				return "/scoped"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newRateLimiterRedis(t)

			var (
				keyGenCalls int
				handlerRuns int
			)

			router := gin.New()
			router.Use(RequestID())
			path := tt.register(
				router,
				RateLimiter(RateLimiterConfig{
					Client: client,
					Limit:  10,
					Window: time.Minute,
					KeyGen: func(*gin.Context) string {
						keyGenCalls++
						return "test-key"
					},
				}),
				func(c *gin.Context) {
					handlerRuns++
					c.Status(http.StatusNoContent)
				},
			)

			requireRateLimitStatus(t, serveRateLimitRequest(t, router, path, nil), http.StatusNoContent)
			if keyGenCalls != 1 {
				t.Errorf("expected KeyGen to run once, ran %d times", keyGenCalls)
			}
			if handlerRuns != 1 {
				t.Errorf("expected handler to run once, ran %d times", handlerRuns)
			}
		})
	}
}

func TestRateLimiter_IsAtomicUnderConcurrency(t *testing.T) {
	const (
		limit    = 10
		requests = 50
	)

	client, _ := newRateLimiterRedis(t)

	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  limit,
			Window: time.Minute,
			KeyGen: func(*gin.Context) string {
				return "client-a"
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	var (
		wg      sync.WaitGroup
		allowed atomic.Int64
	)

	wg.Add(requests)
	for range requests {
		go func() {
			defer wg.Done()

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code == http.StatusNoContent {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != limit {
		t.Errorf("expected exactly %d allowed requests, got %d", limit, got)
	}
}

func TestRateLimiter_AllowsRequestWhenRedisFails(t *testing.T) {
	client, _ := newRateLimiterRedis(t)
	if err := client.Close(); err != nil {
		t.Fatalf("close redis client: %v", err)
	}

	router := gin.New()
	router.Use(
		RequestID(),
		RateLimiter(RateLimiterConfig{
			Client: client,
			Limit:  1,
			Window: time.Minute,
			KeyGen: func(*gin.Context) string {
				return "client-a"
			},
		}),
	)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	requireRateLimitStatus(t, serveRateLimitRequest(t, router, "/test", nil), http.StatusNoContent)
}

func TestRateLimiter_PanicsOnInvalidConfig(t *testing.T) {
	client, _ := newRateLimiterRedis(t)

	valid := RateLimiterConfig{
		Client: client,
		Limit:  1,
		Window: time.Second,
		KeyGen: func(*gin.Context) string {
			return "test-key"
		},
	}

	tests := []struct {
		name   string
		mutate func(*RateLimiterConfig)
		want   string
	}{
		{"nil client", func(c *RateLimiterConfig) { c.Client = nil }, "Redis client"},
		{"zero limit", func(c *RateLimiterConfig) { c.Limit = 0 }, "limit must be greater than zero"},
		{"zero window", func(c *RateLimiterConfig) { c.Window = 0 }, "window must be at least one millisecond"},
		{"nil keygen", func(c *RateLimiterConfig) { c.KeyGen = nil }, "requires a KeyGen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := valid
			tt.mutate(&config)

			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatal("expected RateLimiter to panic")
				}

				message, ok := recovered.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T", recovered)
				}
				if !strings.Contains(message, tt.want) {
					t.Errorf("expected panic message to contain %q, got %q", tt.want, message)
				}
			}()

			RateLimiter(config)
		})
	}
}
