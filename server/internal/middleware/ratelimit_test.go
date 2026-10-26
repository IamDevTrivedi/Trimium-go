package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
			var (
				keyGenCalls int
				handlerRuns int
			)

			handler := func(c *gin.Context) {
				handlerRuns++
				c.Status(http.StatusNoContent)
			}

			router := gin.New()
			path := tt.register(
				router,
				RateLimiter(10, 1_000, func(*gin.Context) string {
					keyGenCalls++
					return "test-key"
				}),
				handler,
			)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

			if w.Code != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
			}
			if keyGenCalls != 1 {
				t.Errorf("expected KeyGen to run once, ran %d times", keyGenCalls)
			}
			if handlerRuns != 1 {
				t.Errorf("expected handler to run once, ran %d times", handlerRuns)
			}
		})
	}
}

func TestRateLimiter_KeyGenRunsPerRequest(t *testing.T) {
	const requests = 3

	var (
		keyGenCalls int
		paths       []string
	)

	router := gin.New()
	router.Use(RateLimiter(100, 60_000, func(c *gin.Context) string {
		keyGenCalls++
		paths = append(paths, c.Request.URL.Path)
		return c.GetHeader("X-API-Key")
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for range requests {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
	}

	if keyGenCalls != requests {
		t.Errorf("expected KeyGen to run %d times, ran %d times", requests, keyGenCalls)
	}
	if len(paths) != requests {
		t.Fatalf("expected %d paths, got %d", requests, len(paths))
	}
	for i, path := range paths {
		if path != "/test" {
			t.Errorf("request %d: KeyGen received path %q, want %q", i, path, "/test")
		}
	}
}
