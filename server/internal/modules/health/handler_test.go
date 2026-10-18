package health

import (
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

var _ Service = (*mockService)(nil)

type mockService struct {
	data map[string]any
	err  error
}

func (m *mockService) GetStatus() (map[string]any, error) {
	return m.data, m.err
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestHandler_HandleHealth(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		err          error
		wantStatus   int
		wantSuccess  bool
		wantMsg      string
		wantMetadata any
	}{
		{
			name:         "returns success envelope with service payload",
			data:         map[string]any{"status": "ok"},
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "Trimium API is healthy",
			wantMetadata: map[string]any{"status": "ok"},
		},
		{
			name:         "nil payload is serialized with null metadata",
			data:         nil,
			wantStatus:   http.StatusOK,
			wantSuccess:  true,
			wantMsg:      "Trimium API is healthy",
			wantMetadata: nil,
		},
		{
			name:         "service error returns error envelope",
			data:         nil,
			err:          errors.New("health check failed"),
			wantStatus:   http.StatusInternalServerError,
			wantSuccess:  false,
			wantMsg:      "health check failed",
			wantMetadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			RegisterRoutes(router, NewHandler(&mockService{data: tt.data, err: tt.err}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
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

func TestRegisterRoutes_OnlyHealthzPathIsRegistered(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router, NewHandler(&mockService{}))

	routes := router.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 registered route, got %d", len(routes))
	}
	if routes[0].Method != http.MethodGet || routes[0].Path != "/healthz" {
		t.Errorf("expected GET /healthz, got %s %s", routes[0].Method, routes[0].Path)
	}
}
