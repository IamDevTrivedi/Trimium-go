package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequestID_SetsUUIDv7InContext(t *testing.T) {
	var handlerID string

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		handlerID = GetRequestID(c)
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if handlerID == "" {
		t.Fatal("expected a request ID to be available to the handler")
	}

	id, err := uuid.Parse(handlerID)
	if err != nil {
		t.Fatalf("request ID %q is not a valid UUID: %v", handlerID, err)
	}
	if got := id.Version(); got != 7 {
		t.Errorf("expected UUID version 7, got %d", got)
	}
	if got := id.Variant(); got != uuid.RFC4122 {
		t.Errorf("expected RFC 4122 UUID variant, got %v", got)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	const requests = 25

	router := gin.New()
	router.Use(RequestID())

	ids := make([]string, 0, requests)
	router.GET("/test", func(c *gin.Context) {
		ids = append(ids, GetRequestID(c))
		c.Status(http.StatusOK)
	})

	for range requests {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
	}

	if len(ids) != requests {
		t.Fatalf("expected %d request IDs, got %d", requests, len(ids))
	}

	seen := make(map[string]struct{}, requests)
	for i, id := range ids {
		if id == "" {
			t.Fatalf("request %d received an empty request ID", i)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("request %d reused request ID %q", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestGetRequestID_Fallbacks(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if got := GetRequestID(c); got != "" {
		t.Errorf("expected empty request ID without middleware, got %q", got)
	}

	c.Set(ReqIDKey, 123)
	if got := GetRequestID(c); got != "" {
		t.Errorf("expected empty request ID for a non-string value, got %q", got)
	}

	c.Set(ReqIDKey, "manual-id")
	if got := GetRequestID(c); got != "manual-id" {
		t.Errorf("expected request ID %q, got %q", "manual-id", got)
	}
}
