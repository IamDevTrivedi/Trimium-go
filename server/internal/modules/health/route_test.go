package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type recordingRateLimiter struct {
	calls []string
}

func (r *recordingRateLimiter) Index(c *gin.Context) {
	r.calls = append(r.calls, "index")
	c.Next()
}

func (r *recordingRateLimiter) Redis(c *gin.Context) {
	r.calls = append(r.calls, "redis")
	c.Next()
}

func (r *recordingRateLimiter) Mongo(c *gin.Context) {
	r.calls = append(r.calls, "mongo")
	c.Next()
}

type recordingHandler struct {
	calls []string
}

func (h *recordingHandler) HandleHealth(c *gin.Context) {
	h.calls = append(h.calls, "health")
	c.Status(http.StatusNoContent)
}

func (h *recordingHandler) HandleRedisHealth(c *gin.Context) {
	h.calls = append(h.calls, "redis")
	c.Status(http.StatusNoContent)
}

func (h *recordingHandler) HandleMongoHealth(c *gin.Context) {
	h.calls = append(h.calls, "mongo")
	c.Status(http.StatusNoContent)
}

func TestRegisterRoutes_WiresMiddlewareAndHandlers(t *testing.T) {
	tests := []struct {
		path     string
		wantRate string
		wantHdlr string
	}{
		{path: "/api/v1/healthz", wantRate: "index", wantHdlr: "health"},
		{path: "/api/v1/healthz/redis", wantRate: "redis", wantHdlr: "redis"},
		{path: "/api/v1/healthz/mongo", wantRate: "mongo", wantHdlr: "mongo"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rl := &recordingRateLimiter{}
			h := &recordingHandler{}
			router := gin.New()
			RegisterRoutes(router, rl, h)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil))

			if w.Code != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
			}
			if len(rl.calls) != 1 || rl.calls[0] != tt.wantRate {
				t.Errorf("expected rate limiter calls %v, got %v", []string{tt.wantRate}, rl.calls)
			}
			if len(h.calls) != 1 || h.calls[0] != tt.wantHdlr {
				t.Errorf("expected handler calls %v, got %v", []string{tt.wantHdlr}, h.calls)
			}
		})
	}
}

func TestRegisterRoutes_RegistersExpectedRoutes(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router, &recordingRateLimiter{}, &recordingHandler{})

	want := map[string]string{
		"/api/v1/healthz":       http.MethodGet,
		"/api/v1/healthz/redis": http.MethodGet,
		"/api/v1/healthz/mongo": http.MethodGet,
	}

	routes := router.Routes()
	if len(routes) != len(want) {
		t.Fatalf("expected %d routes, got %d: %#v", len(want), len(routes), routes)
	}

	for _, route := range routes {
		method, ok := want[route.Path]
		if !ok {
			t.Errorf("unexpected route %s %s", route.Method, route.Path)
			continue
		}
		if route.Method != method {
			t.Errorf("expected %s %s, got %s %s", method, route.Path, route.Method, route.Path)
		}
	}
}
