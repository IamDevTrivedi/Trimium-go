package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

const (
	chromeWindowsUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	iphoneSafariUA  = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	ipadSafariUA    = "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	googlebotUA     = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

func TestParseUA(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want UserAgentInfo
	}{
		{
			name: "chrome on windows",
			raw:  chromeWindowsUA,
			want: UserAgentInfo{
				Raw:            chromeWindowsUA,
				Browser:        "Chrome",
				BrowserVersion: "120.0.0.0",
				OS:             "Windows",
				OSVersion:      "10.0",
				DeviceType:     "desktop",
			},
		},
		{
			name: "safari on iphone",
			raw:  iphoneSafariUA,
			want: UserAgentInfo{
				Raw:            iphoneSafariUA,
				Browser:        "Safari",
				BrowserVersion: "17.0",
				OS:             "iOS",
				OSVersion:      "17.0",
				Device:         "iPhone",
				DeviceType:     "mobile",
			},
		},
		{
			name: "safari on ipad",
			raw:  ipadSafariUA,
			want: UserAgentInfo{
				Raw:            ipadSafariUA,
				Browser:        "Safari",
				BrowserVersion: "17.0",
				OS:             "iOS",
				OSVersion:      "17.0",
				Device:         "iPad",
				DeviceType:     "tablet",
			},
		},
		{
			name: "googlebot",
			raw:  googlebotUA,
			want: UserAgentInfo{
				Raw:            googlebotUA,
				Browser:        "Googlebot",
				BrowserVersion: "2.1",
				DeviceType:     "bot",
				IsBot:          true,
			},
		},
		{
			name: "empty header",
			raw:  "",
			want: UserAgentInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUA(tt.raw)
			if got == nil {
				t.Fatal("parseUA() returned nil")
			}
			if *got != tt.want {
				t.Errorf("parseUA(%q) = %+v, want %+v", tt.raw, *got, tt.want)
			}
		})
	}
}

func TestDeviceType(t *testing.T) {
	tests := []struct {
		name string
		ua   useragent.UserAgent
		want string
	}{
		{"bot", useragent.UserAgent{Bot: true}, "bot"},
		{"bot wins over other flags", useragent.UserAgent{Bot: true, Tablet: true, Mobile: true, Desktop: true}, "bot"},
		{"tablet", useragent.UserAgent{Tablet: true}, "tablet"},
		{"tablet wins over mobile", useragent.UserAgent{Tablet: true, Mobile: true}, "tablet"},
		{"mobile", useragent.UserAgent{Mobile: true}, "mobile"},
		{"desktop", useragent.UserAgent{Desktop: true}, "desktop"},
		{"unknown", useragent.UserAgent{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceType(tt.ua); got != tt.want {
				t.Errorf("deviceType(%+v) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func TestRequestUserAgent_SetsParsedHeader(t *testing.T) {
	var (
		got        *UserAgentInfo
		nextCalled bool
	)

	router := gin.New()
	router.Use(RequestUserAgent())
	router.GET("/test", func(c *gin.Context) {
		nextCalled = true
		got = GetRequestUA(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", chromeWindowsUA)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !nextCalled {
		t.Fatal("expected the next handler to run")
	}
	if got == nil {
		t.Fatal("GetRequestUA() returned nil")
	}
	if got.Raw != chromeWindowsUA {
		t.Errorf("expected raw user agent %q, got %q", chromeWindowsUA, got.Raw)
	}
	if got.Browser != "Chrome" {
		t.Errorf("expected browser Chrome, got %q", got.Browser)
	}
	if got.DeviceType != "desktop" {
		t.Errorf("expected device type desktop, got %q", got.DeviceType)
	}
}

func TestRequestUserAgent_EmptyHeader(t *testing.T) {
	var got *UserAgentInfo

	router := gin.New()
	router.Use(RequestUserAgent())
	router.GET("/test", func(c *gin.Context) {
		got = GetRequestUA(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

	if got == nil {
		t.Fatal("GetRequestUA() returned nil for an empty User-Agent")
	}
	if *got != (UserAgentInfo{}) {
		t.Errorf("expected empty UserAgentInfo, got %+v", *got)
	}
}

func TestGetRequestUA_Fallbacks(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	got := GetRequestUA(c)
	if got == nil {
		t.Fatal("GetRequestUA() returned nil without middleware")
	}
	if *got != (UserAgentInfo{}) {
		t.Errorf("expected empty UserAgentInfo without middleware, got %+v", *got)
	}

	c.Set(ReqUAKey, "not-a-user-agent")
	got = GetRequestUA(c)
	if got == nil {
		t.Fatal("GetRequestUA() returned nil for a non-struct value")
	}
	if *got != (UserAgentInfo{}) {
		t.Errorf("expected empty UserAgentInfo for a non-struct value, got %+v", *got)
	}
}
