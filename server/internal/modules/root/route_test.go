package root

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type recordingRateLimiter struct {
	calls int
}

func (r *recordingRateLimiter) Index(c *gin.Context) {
	r.calls++
	c.Next()
}

type recordingHandler struct {
	calls int
}

func (h *recordingHandler) HandleRoot(c *gin.Context) {
	h.calls++
	c.Status(http.StatusNoContent)
}

func TestRegisterRoutes_WiresMiddlewareAndHandler(t *testing.T) {
	rl := &recordingRateLimiter{}
	h := &recordingHandler{}
	router := gin.New()
	RegisterRoutes(router, rl, h)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
	if rl.calls != 1 {
		t.Errorf("expected rate limiter to run once, ran %d times", rl.calls)
	}
	if h.calls != 1 {
		t.Errorf("expected handler to run once, ran %d times", h.calls)
	}
}

func TestRegisterRoutes_RegistersSingleRootRoute(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router, &recordingRateLimiter{}, &recordingHandler{})

	routes := router.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d: %#v", len(routes), routes)
	}
	if routes[0].Method != http.MethodGet || routes[0].Path != "/api/v1" {
		t.Errorf("expected GET /api/v1, got %s %s", routes[0].Method, routes[0].Path)
	}
}
