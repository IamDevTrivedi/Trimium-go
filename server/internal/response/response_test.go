package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// serve runs write as the handler for GET /test and returns the recorded response.
func serve(write func(c *gin.Context)) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/test", write)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) Response {
	t.Helper()

	var got Response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	return got
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name     string
		success  bool
		msg      string
		metadata any
	}{
		{
			name:     "success with map metadata",
			success:  true,
			msg:      "welcome",
			metadata: map[string]any{"message": "hello", "status": "ok"},
		},
		{
			name:     "error with nil metadata",
			success:  false,
			msg:      "boom",
			metadata: nil,
		},
		{
			name:     "empty message",
			success:  true,
			msg:      "",
			metadata: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			got := NewResponse(tt.success, tt.msg, tt.metadata)
			after := time.Now().UnixMilli()

			if got.Success != tt.success {
				t.Errorf("expected success %v, got %v", tt.success, got.Success)
			}
			if got.Msg != tt.msg {
				t.Errorf("expected msg %q, got %q", tt.msg, got.Msg)
			}
			if !reflect.DeepEqual(got.Metadata, tt.metadata) {
				t.Errorf("expected metadata %v, got %v", tt.metadata, got.Metadata)
			}
			if got.Timestamp < before || got.Timestamp > after {
				t.Errorf("expected timestamp in [%d, %d], got %d", before, after, got.Timestamp)
			}
		})
	}
}

func TestWriters(t *testing.T) {
	payload := map[string]any{"id": "user-1", "name": "Trimium"}

	tests := []struct {
		name         string
		write        func(c *gin.Context)
		wantStatus   int
		wantSuccess  bool
		wantMsg      string
		wantMetadata any
	}{
		{
			name:         "Success writes 200 with payload",
			write:        func(c *gin.Context) { Success(c, "all good", payload) },
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "all good",
			wantMetadata: payload,
		},
		{
			name:         "Success writes null metadata for nil payload",
			write:        func(c *gin.Context) { Success(c, "no payload", nil) },
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "no payload",
			wantMetadata: nil,
		},
		{
			name:         "Created writes 201 with payload",
			write:        func(c *gin.Context) { Created(c, "created", payload) },
			wantStatus:   http.StatusCreated,
			wantSuccess:  true,
			wantMsg:      "created",
			wantMetadata: payload,
		},
		{
			name:         "Error writes status with nil metadata",
			write:        func(c *gin.Context) { Error(c, http.StatusBadRequest, "invalid input") },
			wantStatus:   http.StatusBadRequest,
			wantSuccess:  false,
			wantMsg:      "invalid input",
			wantMetadata: nil,
		},
		{
			name:         "Error supports 500",
			write:        func(c *gin.Context) { Error(c, http.StatusInternalServerError, "internal error") },
			wantStatus:   http.StatusInternalServerError,
			wantSuccess:  false,
			wantMsg:      "internal error",
			wantMetadata: nil,
		},
		{
			name:         "JSON writes custom status and success flag",
			write:        func(c *gin.Context) { JSON(c, http.StatusTeapot, false, "teapot", payload) },
			wantStatus:   http.StatusTeapot,
			wantSuccess:  false,
			wantMsg:      "teapot",
			wantMetadata: payload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(tt.write)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", ct)
			}

			got := decodeResponse(t, w)
			if got.Success != tt.wantSuccess {
				t.Errorf("expected success %v, got %v", tt.wantSuccess, got.Success)
			}
			if got.Msg != tt.wantMsg {
				t.Errorf("expected msg %q, got %q", tt.wantMsg, got.Msg)
			}
			if !reflect.DeepEqual(got.Metadata, tt.wantMetadata) {
				t.Errorf("expected metadata %v, got %v", tt.wantMetadata, got.Metadata)
			}
			if got.Timestamp <= 0 {
				t.Errorf("expected positive timestamp, got %d", got.Timestamp)
			}
		})
	}
}

func TestError_SerializesNullMetadata(t *testing.T) {
	w := serve(func(c *gin.Context) { Error(c, http.StatusNotFound, "not found") })

	// Pin the wire format: the metadata key must always be present,
	// even when nil, so clients can rely on a uniform envelope.
	if !strings.Contains(w.Body.String(), `"metadata":null`) {
		t.Errorf("expected body to contain %q, got %s", `"metadata":null`, w.Body.String())
	}
}
