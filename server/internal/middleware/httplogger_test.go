package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"log/slog"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// newTestRouter wires the middleware with a buffer-backed JSON logger so
// assertions can read the emitted entries directly.
func newTestRouter(buf *bytes.Buffer, handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(requestLogger(slog.New(slog.NewJSONHandler(buf, nil))))
	router.GET("/test", handler)
	return router
}

func parseLogEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("failed to parse log line %q: %v", line, err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestRequestLogger_LogsRequestMetadata(t *testing.T) {
	var buf bytes.Buffer
	router := newTestRouter(&buf, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test?foo=bar&baz=1", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	entries := parseLogEntries(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log entry, got %d", len(entries))
	}

	entry := entries[0]

	if entry["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", entry["level"])
	}
	if entry["msg"] != "http request" {
		t.Errorf("expected msg %q, got %v", "http request", entry["msg"])
	}
	if entry["method"] != "GET" {
		t.Errorf("expected method GET, got %v", entry["method"])
	}
	if entry["path"] != "/test" {
		t.Errorf("expected path /test, got %v", entry["path"])
	}
	if entry["query"] != "foo=bar&baz=1" {
		t.Errorf("expected query %q, got %v", "foo=bar&baz=1", entry["query"])
	}
	if entry["status"] != float64(http.StatusOK) {
		t.Errorf("expected status 200, got %v", entry["status"])
	}
	if entry["user_agent"] != "test-agent/1.0" {
		t.Errorf("expected user agent %q, got %v", "test-agent/1.0", entry["user_agent"])
	}
	if ip, ok := entry["client_ip"].(string); !ok || ip == "" {
		t.Errorf("expected non-empty client_ip, got %v", entry["client_ip"])
	}
	if latency, ok := entry["latency_ms"].(float64); !ok || latency < 0 {
		t.Errorf("expected non-negative latency_ms, got %v", entry["latency_ms"])
	}
	if _, ok := entry["errors"]; ok {
		t.Errorf("did not expect errors attribute, got %v", entry["errors"])
	}
}

func TestRequestLogger_LevelByStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		pushError bool
		wantLevel string
	}{
		{"200 logs INFO", http.StatusOK, false, "INFO"},
		{"201 logs INFO", http.StatusCreated, false, "INFO"},
		{"404 logs WARN", http.StatusNotFound, false, "WARN"},
		{"400 logs WARN", http.StatusBadRequest, false, "WARN"},
		{"500 logs ERROR", http.StatusInternalServerError, false, "ERROR"},
		{"503 logs ERROR", http.StatusServiceUnavailable, false, "ERROR"},
		{"gin error is attached to the entry", http.StatusInternalServerError, true, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			status := tt.status
			pushError := tt.pushError

			router := newTestRouter(&buf, func(c *gin.Context) {
				if pushError {
					_ = c.Error(errors.New("boom"))
				}
				c.Status(status)
			})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

			entries := parseLogEntries(t, &buf)
			if len(entries) != 1 {
				t.Fatalf("expected exactly 1 log entry, got %d", len(entries))
			}

			entry := entries[0]
			if entry["level"] != tt.wantLevel {
				t.Errorf("expected level %s, got %v", tt.wantLevel, entry["level"])
			}
			if entry["status"] != float64(status) {
				t.Errorf("expected status %d, got %v", status, entry["status"])
			}

			errs, hasErrors := entry["errors"]
			if pushError {
				if !hasErrors {
					t.Fatal("expected errors attribute, got none")
				}
				if !strings.Contains(fmt.Sprint(errs), "boom") {
					t.Errorf("expected errors to contain %q, got %v", "boom", errs)
				}
			} else if hasErrors {
				t.Errorf("did not expect errors attribute, got %v", errs)
			}
		})
	}
}

func TestRequestLogger_UnsetStatusDefaultsTo200(t *testing.T) {
	var buf bytes.Buffer
	router := newTestRouter(&buf, func(_ *gin.Context) {})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

	entries := parseLogEntries(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log entry, got %d", len(entries))
	}
	if entries[0]["status"] != float64(http.StatusOK) {
		t.Errorf("expected status 200, got %v", entries[0]["status"])
	}
}
