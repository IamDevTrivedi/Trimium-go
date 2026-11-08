package root

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func (r *recordingRateLimiter) WelcomeEmail(c *gin.Context) {
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

func (h *recordingHandler) HandleSendWelcomeEmail(c *gin.Context) {
	h.calls++
	c.Status(http.StatusNoContent)
}

func TestRegisterRoutes_WiresMiddlewareAndHandler(t *testing.T) {
	rl := &recordingRateLimiter{}
	h := &recordingHandler{}
	router := gin.New()
	RegisterRoutes(router, rl.Index, rl.WelcomeEmail, h)

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

func TestRegisterRoutes_WiresWelcomeEmailRoute(t *testing.T) {
	rl := &recordingRateLimiter{}
	h := &recordingHandler{}
	router := gin.New()
	RegisterRoutes(router, rl.Index, rl.WelcomeEmail, h)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, "/api/v1/emails/welcome", strings.NewReader(`{"to":["user@example.com"]}`),
	))

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

func TestRegisterRoutes_RegistersAllRoutes(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router, (&recordingRateLimiter{}).Index, (&recordingRateLimiter{}).WelcomeEmail, &recordingHandler{})

	routes := router.Routes()
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d: %#v", len(routes), routes)
	}

	want := map[string]bool{
		"GET /api/v1":                 false,
		"POST /api/v1/emails/welcome": false,
	}
	for _, r := range routes {
		key := r.Method + " " + r.Path
		if _, ok := want[key]; !ok {
			t.Errorf("unexpected route %s", key)
			continue
		}
		want[key] = true
	}
	for key, found := range want {
		if !found {
			t.Errorf("expected route %s to be registered", key)
		}
	}
}
