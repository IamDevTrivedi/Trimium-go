package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"server/internal/middleware"
	"server/internal/response"
)

func TestRegister(t *testing.T) {
	store, _ := newTestRedis(t)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestIP())

	// The mongo dependency is only exercised by the mongo route, so nil is
	// sufficient for wiring the app and redis routes.
	Register(router, store, nil)

	if got := len(router.Routes()); got != 3 {
		t.Fatalf("expected 3 registered routes, got %d", got)
	}

	tests := []struct {
		path string
		msg  string
	}{
		{path: "/api/v1/healthz", msg: "App is working"},
		{path: "/api/v1/healthz/redis", msg: "Redis is working"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			req.RemoteAddr = "203.0.113.40:1234"
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
			}

			var got response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}
			if !got.Success {
				t.Error("expected success to be true")
			}
			if got.Msg != tt.msg {
				t.Errorf("expected msg %q, got %q", tt.msg, got.Msg)
			}
		})
	}
}
