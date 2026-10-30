package root

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

	Register(router, store, nil)

	if got := len(router.Routes()); got != 1 {
		t.Fatalf("expected 1 registered route, got %d", got)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1", nil)
	req.RemoteAddr = "203.0.113.70:1234"
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
	if got.Msg != "Welcome to Trimium API" {
		t.Errorf("expected msg %q, got %q", "Welcome to Trimium API", got.Msg)
	}

	metadata, ok := got.Metadata.(map[string]any)
	if !ok {
		t.Fatalf("expected object metadata, got %T", got.Metadata)
	}
	if metadata["name"] != "Trimium API" {
		t.Errorf("expected name %q, got %v", "Trimium API", metadata["name"])
	}
	if metadata["status"] != "ok" {
		t.Errorf("expected status %q, got %v", "ok", metadata["status"])
	}
	if version, _ := metadata["version"].(string); version == "" {
		t.Error("expected a non-empty version")
	}
}
