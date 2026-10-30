package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

type mockService struct {
	appStatus   *StatusResponse
	appErr      error
	redisStatus *StatusResponse
	redisErr    error
	mongoStatus *StatusResponse
	mongoErr    error
}

func (m *mockService) GetAppStatus(context.Context) (*StatusResponse, error) {
	return m.appStatus, m.appErr
}

func (m *mockService) GetRedisStatus(context.Context) (*StatusResponse, error) {
	return m.redisStatus, m.redisErr
}

func (m *mockService) GetMongoStatus(context.Context) (*StatusResponse, error) {
	return m.mongoStatus, m.mongoErr
}

func TestHandler(t *testing.T) {
	ok := &StatusResponse{Status: "ok"}

	tests := []struct {
		name         string
		register     func(*gin.Engine, Handler)
		service      *mockService
		wantStatus   int
		wantSuccess  bool
		wantMsg      string
		wantMetadata any
	}{
		{
			name:         "health returns success envelope",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleHealth) },
			service:      &mockService{appStatus: ok},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "App is working",
			wantMetadata: map[string]any{"status": "ok"},
		},
		{
			name:         "health serializes nil payload as null metadata",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleHealth) },
			service:      &mockService{},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "App is working",
			wantMetadata: nil,
		},
		{
			name:         "health returns error envelope when service fails",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleHealth) },
			service:      &mockService{appErr: errors.New("boom")},
			wantStatus:   http.StatusInternalServerError,
			wantSuccess:  false,
			wantMsg:      "App is not working",
			wantMetadata: nil,
		},
		{
			name:         "redis returns success envelope",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleRedisHealth) },
			service:      &mockService{redisStatus: ok},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "Redis is working",
			wantMetadata: map[string]any{"status": "ok"},
		},
		{
			name:         "redis returns 503 when service fails",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleRedisHealth) },
			service:      &mockService{redisErr: errors.New("connection refused")},
			wantStatus:   http.StatusServiceUnavailable,
			wantSuccess:  false,
			wantMsg:      "Redis is not working",
			wantMetadata: nil,
		},
		{
			name:         "mongo returns success envelope",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleMongoHealth) },
			service:      &mockService{mongoStatus: ok},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "Mongo is working",
			wantMetadata: map[string]any{"status": "ok"},
		},
		{
			name:         "mongo returns 503 when service fails",
			register:     func(r *gin.Engine, h Handler) { r.GET("/test", h.HandleMongoHealth) },
			service:      &mockService{mongoErr: errors.New("server selection timeout")},
			wantStatus:   http.StatusServiceUnavailable,
			wantSuccess:  false,
			wantMsg:      "Mongo is not working",
			wantMetadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			tt.register(router, NewHandler(tt.service))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", ct)
			}

			var got response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}

			if got.Success != tt.wantSuccess {
				t.Errorf("expected success %v, got %v", tt.wantSuccess, got.Success)
			}
			if got.Msg != tt.wantMsg {
				t.Errorf("expected msg %q, got %q", tt.wantMsg, got.Msg)
			}
			if got.Timestamp <= 0 {
				t.Errorf("expected positive timestamp, got %d", got.Timestamp)
			}
			if !reflect.DeepEqual(got.Metadata, tt.wantMetadata) {
				t.Errorf("expected metadata %v, got %v", tt.wantMetadata, got.Metadata)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	if h := NewHandler(&mockService{}); h == nil {
		t.Fatal("expected NewHandler to return a non-nil handler")
	}
}
