package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResolveIP(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"public IPv4", "8.8.8.8", "8.8.8.8"},
		{"documentation IPv4", "203.0.113.9", "203.0.113.9"},
		{"public IPv6", "2001:db8::1", "2001:db8::1"},
		{"IPv4-mapped IPv6", "::ffff:192.0.2.33", "192.0.2.33"},
		{"loopback IPv4", "127.0.0.1", localhostFallbackIP},
		{"loopback IPv4 range", "127.9.8.7", localhostFallbackIP},
		{"loopback IPv6", "::1", localhostFallbackIP},
		{"unspecified IPv4", "0.0.0.0", localhostFallbackIP},
		{"unspecified IPv6", "::", localhostFallbackIP},
		{"invalid input", "not-an-ip", localhostFallbackIP},
		{"empty input", "", localhostFallbackIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveIP(tt.raw); got != tt.want {
				t.Errorf("resolveIP(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestRequestIP_SetsResolvedIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{"public IPv4", "203.0.113.7:4321", "203.0.113.7"},
		{"public IPv6", "[2001:db8::1]:4321", "2001:db8::1"},
		{"loopback", "127.0.0.1:4321", localhostFallbackIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got        string
				nextCalled bool
			)

			router := gin.New()
			router.Use(RequestIP())
			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				got = GetRequestIP(c)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if !nextCalled {
				t.Fatal("expected the next handler to run")
			}
			if got != tt.want {
				t.Errorf("GetRequestIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetRequestIP_Fallbacks(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if got := GetRequestIP(c); got != "" {
		t.Errorf("expected empty IP without middleware, got %q", got)
	}

	c.Set(ReqIPKey, 42)
	if got := GetRequestIP(c); got != "" {
		t.Errorf("expected empty IP for a non-string value, got %q", got)
	}

	c.Set(ReqIPKey, "198.51.100.4")
	if got := GetRequestIP(c); got != "198.51.100.4" {
		t.Errorf("expected IP %q, got %q", "198.51.100.4", got)
	}
}
