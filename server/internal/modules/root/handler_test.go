package root

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockRootService struct {
	mockStatus map[string]any
}

func (m *mockRootService) GetStatus() map[string]any {
	return m.mockStatus
}

func TestHandler_HandleRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockData     map[string]any
		expectedCode int
	}{
		{
			name: "returns status 200 with service payload",
			mockData: map[string]any{
				"message": "Custom Welcome Message",
				"status":  "healthy",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockRootService{mockStatus: tc.mockData}
			h := NewHandler(mockSvc)

			router := gin.New()
			RegisterRoutes(router, h)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, w.Code)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}

			if body["message"] != tc.mockData["message"] {
				t.Errorf("expected message %v, got %v", tc.mockData["message"], body["message"])
			}
			if body["status"] != tc.mockData["status"] {
				t.Errorf("expected status %v, got %v", tc.mockData["status"], body["status"])
			}
		})
	}
}
