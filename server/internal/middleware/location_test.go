package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// geoDBTestPath points at the development GeoLite2 database, relative to this
// package directory. The database is gitignored, so callers skip when it is
// absent.
func geoDBTestPath(t *testing.T) string {
	t.Helper()

	dbPath := filepath.Join("..", "..", "data", "geoip", "GeoLite2-City.mmdb")
	if _, err := os.Stat(dbPath); err != nil {
		t.Skipf("GeoLite2 database unavailable at %s: %v", dbPath, err)
	}
	return dbPath
}

func openGeoTestResolver(t *testing.T) *GeoResolver {
	t.Helper()

	g, err := NewGeoResolver(geoDBTestPath(t))
	if err != nil {
		t.Fatalf("NewGeoResolver() error: %v", err)
	}
	t.Cleanup(func() {
		if err := g.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	})
	return g
}

func TestNewGeoResolver_InvalidPath(t *testing.T) {
	g, err := NewGeoResolver(filepath.Join(t.TempDir(), "missing.mmdb"))
	if err == nil {
		if g != nil {
			_ = g.Close()
		}
		t.Fatal("expected an error for a missing database")
	}
	if g != nil {
		t.Errorf("expected a nil resolver on error, got %+v", g)
	}
}

func TestGeoResolver_OpenAndClose(t *testing.T) {
	g, err := NewGeoResolver(geoDBTestPath(t))
	if err != nil {
		t.Fatalf("NewGeoResolver() error: %v", err)
	}
	if g.db == nil {
		t.Fatal("expected an open maxmind database")
	}
	if err := g.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestGeoResolver_Lookup(t *testing.T) {
	g := openGeoTestResolver(t)

	tests := []struct {
		name            string
		ip              string
		wantCountry     string
		wantCountryName string
		wantContinent   string
		wantTimezone    string
		wantEU          bool
	}{
		{
			name:            "MaxMind London test IP",
			ip:              "81.2.69.142",
			wantCountry:     "GB",
			wantCountryName: "United Kingdom",
			wantContinent:   "EU",
			wantTimezone:    "Europe/London",
			wantEU:          false,
		},
		{
			name:            "Swedish IP",
			ip:              "89.160.20.112",
			wantCountry:     "SE",
			wantCountryName: "Sweden",
			wantContinent:   "EU",
			wantTimezone:    "Europe/Stockholm",
			wantEU:          true,
		},
		{
			name:            "Google public DNS",
			ip:              "8.8.8.8",
			wantCountry:     "US",
			wantCountryName: "United States",
			wantContinent:   "NA",
			wantTimezone:    "America/Chicago",
			wantEU:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := g.lookup(tt.ip)
			if info == nil {
				t.Fatal("lookup() returned nil")
			}

			if info.Country != tt.wantCountry {
				t.Errorf("Country = %q, want %q", info.Country, tt.wantCountry)
			}
			if info.CountryName != tt.wantCountryName {
				t.Errorf("CountryName = %q, want %q", info.CountryName, tt.wantCountryName)
			}
			if info.Continent != tt.wantContinent {
				t.Errorf("Continent = %q, want %q", info.Continent, tt.wantContinent)
			}
			if info.Timezone != tt.wantTimezone {
				t.Errorf("Timezone = %q, want %q", info.Timezone, tt.wantTimezone)
			}
			if info.IsEU != tt.wantEU {
				t.Errorf("IsEU = %v, want %v", info.IsEU, tt.wantEU)
			}
		})
	}
}

func TestGeoResolver_LookupLondonDetails(t *testing.T) {
	g := openGeoTestResolver(t)

	info := g.lookup("81.2.69.142")
	if info == nil {
		t.Fatal("lookup() returned nil")
	}

	if info.Region != "England" {
		t.Errorf("Region = %q, want %q", info.Region, "England")
	}
	if info.RegionCode != "ENG" {
		t.Errorf("RegionCode = %q, want %q", info.RegionCode, "ENG")
	}
	if info.City == "" {
		t.Error("expected a non-empty City")
	}
	if info.PostalCode == "" {
		t.Error("expected a non-empty PostalCode")
	}
	if info.Latitude < 51 || info.Latitude > 52 {
		t.Errorf("Latitude = %v, want a value around London (51-52)", info.Latitude)
	}
	if info.Longitude < -1 || info.Longitude > 0 {
		t.Errorf("Longitude = %v, want a value around London (-1-0)", info.Longitude)
	}
	if info.AccuracyKm == 0 {
		t.Error("expected a non-zero AccuracyKm")
	}
}

func TestGeoResolver_LookupInvalidIP(t *testing.T) {
	g := openGeoTestResolver(t)

	for _, raw := range []string{"", "not-an-ip", "999.1.1.1", "localhost"} {
		t.Run(raw, func(t *testing.T) {
			info := g.lookup(raw)
			if info == nil {
				t.Fatal("lookup() returned nil")
			}
			if *info != (GeoInfo{}) {
				t.Errorf("lookup(%q) = %+v, want an empty GeoInfo", raw, *info)
			}
		})
	}
}

func TestGeoResolver_LookupReservedIP(t *testing.T) {
	g := openGeoTestResolver(t)

	info := g.lookup("192.0.2.1")
	if info == nil {
		t.Fatal("lookup() returned nil")
	}
	if *info != (GeoInfo{}) {
		t.Errorf("lookup() for a reserved IP = %+v, want an empty GeoInfo", *info)
	}
}

func TestRequestGeo_SetsGeoInfo(t *testing.T) {
	g := openGeoTestResolver(t)

	var (
		got        *GeoInfo
		stored     any
		hasGeoKey  bool
		nextCalled bool
	)

	router := gin.New()
	router.Use(RequestIP(), RequestGeo(g))
	router.GET("/test", func(c *gin.Context) {
		nextCalled = true
		got = GetRequestGeo(c)
		stored, hasGeoKey = c.Get(ReqGeoKey)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	req.RemoteAddr = "81.2.69.142:5555"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !nextCalled {
		t.Fatal("expected the next handler to run")
	}
	if !hasGeoKey {
		t.Fatal("expected the geo info to be stored in the context")
	}
	if _, ok := stored.(*GeoInfo); !ok {
		t.Fatalf("expected a *GeoInfo in the context, got %T", stored)
	}
	if got == nil {
		t.Fatal("GetRequestGeo() returned nil")
	}
	if got.Country != "GB" {
		t.Errorf("Country = %q, want %q", got.Country, "GB")
	}
	if got.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", got.Timezone, "Europe/London")
	}
}

func TestRequestGeo_UnknownIPStillSetsGeoInfo(t *testing.T) {
	g := openGeoTestResolver(t)

	var got *GeoInfo

	router := gin.New()
	router.Use(RequestIP(), RequestGeo(g))
	router.GET("/test", func(c *gin.Context) {
		got = GetRequestGeo(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.0.2.1:5555"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got == nil {
		t.Fatal("GetRequestGeo() returned nil")
	}
	if *got != (GeoInfo{}) {
		t.Errorf("expected an empty GeoInfo for an unknown IP, got %+v", *got)
	}
}

func TestGetRequestGeo_Fallbacks(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	got := GetRequestGeo(c)
	if got == nil {
		t.Fatal("GetRequestGeo() returned nil without middleware")
	}
	if *got != (GeoInfo{}) {
		t.Errorf("expected empty GeoInfo without middleware, got %+v", *got)
	}

	c.Set(ReqGeoKey, "not-geo-info")
	got = GetRequestGeo(c)
	if got == nil {
		t.Fatal("GetRequestGeo() returned nil for a non-struct value")
	}
	if *got != (GeoInfo{}) {
		t.Errorf("expected empty GeoInfo for a non-struct value, got %+v", *got)
	}
}
