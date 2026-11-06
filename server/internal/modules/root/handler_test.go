package root

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

type mockService struct {
	info         *AppInfo
	err          error
	welcomeErr   error
	recipients   []string
	welcomeCalls int
}

func (m *mockService) GetAppInfo(context.Context) (*AppInfo, error) {
	return m.info, m.err
}

func (m *mockService) SendWelcomeEmail(_ context.Context, to []string) error {
	m.welcomeCalls++
	m.recipients = to
	return m.welcomeErr
}

func TestHandler_HandleRoot(t *testing.T) {
	tests := []struct {
		name         string
		service      *mockService
		wantStatus   int
		wantSuccess  bool
		wantMsg      string
		wantMetadata any
	}{
		{
			name: "returns success envelope with app info",
			service: &mockService{info: &AppInfo{
				Name:    "Trimium API",
				Version: "v1.2.3",
				Status:  "ok",
			}},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantMsg:     "Welcome to Trimium API",
			wantMetadata: map[string]any{
				"name":    "Trimium API",
				"version": "v1.2.3",
				"status":  "ok",
			},
		},
		{
			name:         "nil payload is serialized as null metadata",
			service:      &mockService{},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "Welcome to Trimium API",
			wantMetadata: nil,
		},
		{
			name:         "service error returns error envelope",
			service:      &mockService{err: errors.New("boom")},
			wantStatus:   http.StatusInternalServerError,
			wantSuccess:  false,
			wantMsg:      "App is not working",
			wantMetadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/api/v1", NewHandler(tt.service).HandleRoot)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1", nil))

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

func TestHandler_HandleSendWelcomeEmail(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		service     *mockService
		wantStatus  int
		wantSuccess bool
		wantMsg     string
		wantTo      []string
	}{
		{
			name:        "queues welcome email for valid recipients",
			body:        `{"to":["user@example.com"]}`,
			service:     &mockService{},
			wantStatus:  http.StatusAccepted,
			wantSuccess: true,
			wantMsg:     "Welcome email queued",
			wantTo:      []string{"user@example.com"},
		},
		{
			name:        "queues for multiple recipients",
			body:        `{"to":["a@example.com","b@example.com"]}`,
			service:     &mockService{},
			wantStatus:  http.StatusAccepted,
			wantSuccess: true,
			wantMsg:     "Welcome email queued",
			wantTo:      []string{"a@example.com", "b@example.com"},
		},
		{
			name:        "malformed json returns bad request",
			body:        `{"to":`,
			service:     &mockService{},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
			wantMsg:     "Invalid request body",
		},
		{
			name:        "empty recipient list returns bad request",
			body:        `{"to":[]}`,
			service:     &mockService{},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
			wantMsg:     "Invalid request body",
		},
		{
			name:        "invalid recipient address returns bad request",
			body:        `{"to":["not-an-email"]}`,
			service:     &mockService{},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
			wantMsg:     "Invalid request body",
		},
		{
			name:        "service error returns error envelope",
			body:        `{"to":["user@example.com"]}`,
			service:     &mockService{welcomeErr: errors.New("queue down")},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
			wantMsg:     "Could not queue welcome email",
			wantTo:      []string{"user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/api/v1/emails/welcome", NewHandler(tt.service).HandleSendWelcomeEmail)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequestWithContext(
				t.Context(), http.MethodPost, "/api/v1/emails/welcome", strings.NewReader(tt.body),
			))

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
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

			if tt.wantTo != nil || tt.service.welcomeCalls > 0 {
				if tt.service.welcomeCalls != 1 {
					t.Errorf("expected service to be called once, called %d times", tt.service.welcomeCalls)
				}
				if !reflect.DeepEqual(tt.service.recipients, tt.wantTo) {
					t.Errorf("expected recipients %v, got %v", tt.wantTo, tt.service.recipients)
				}
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	if h := NewHandler(&mockService{}); h == nil {
		t.Fatal("expected NewHandler to return a non-nil handler")
	}
}
