package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestNewGeoResolver_InvalidPath verifies that opening a non-existent mmdb
// file returns a non-nil error and a nil resolver.
func TestNewGeoResolver_InvalidPath(t *testing.T) {
	resolver, err := NewGeoResolver(GeoResolverConfig{DBPath: "/nonexistent/path/to/GeoLite2-City.mmdb"})
	if err == nil {
		t.Fatal("expected an error for non-existent mmdb path, got nil")
	}
	if resolver != nil {
		t.Errorf("expected nil resolver on error, got %+v", resolver)
	}
}

// TestGetRequestGeo_NoMiddleware verifies that GetRequestGeo never returns nil
// when RequestGeo middleware was not installed.
func TestGetRequestGeo_NoMiddleware(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		info := GetRequestGeo(c)
		if info == nil {
			t.Error("GetRequestGeo returned nil without middleware")
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
}

// TestGetRequestGeo_WrongType verifies that a non-*GeoInfo value stored at
// ReqGeoKey is ignored and an empty GeoInfo is returned instead of panicking.
func TestGetRequestGeo_WrongType(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// Store an integer (wrong type) at the geo key.
		c.Set(ReqGeoKey, 42)

		info := GetRequestGeo(c)
		if info == nil {
			t.Error("GetRequestGeo returned nil for wrong-type value")
			return
		}
		// All fields must be zero values — no panic, no garbage data.
		if info.Country != "" || info.City != "" || info.Timezone != "" {
			t.Errorf("expected empty GeoInfo for wrong-type value, got %+v", info)
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
}

// TestGetRequestGeo_ValidValue verifies that a correctly typed *GeoInfo stored
// at ReqGeoKey is returned as-is by GetRequestGeo.
func TestGetRequestGeo_ValidValue(t *testing.T) {
	want := &GeoInfo{
		Country:     "US",
		CountryName: "United States",
		City:        "New York",
		Timezone:    "America/New_York",
		IsEU:        false,
	}

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set(ReqGeoKey, want)

		got := GetRequestGeo(c)
		if got == nil {
			t.Error("GetRequestGeo returned nil")
			return
		}
		if got != want {
			t.Errorf("GetRequestGeo returned a different pointer: got %p, want %p", got, want)
		}
		if got.Country != want.Country {
			t.Errorf("Country = %q, want %q", got.Country, want.Country)
		}
		if got.City != want.City {
			t.Errorf("City = %q, want %q", got.City, want.City)
		}
		if got.Timezone != want.Timezone {
			t.Errorf("Timezone = %q, want %q", got.Timezone, want.Timezone)
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
}

// TestGeoResolver_Lookup_InvalidIP verifies that an unparseable IP returns a
// non-nil, zero-value GeoInfo rather than panicking or returning nil.
func TestGeoResolver_Lookup_InvalidIP(t *testing.T) {
	// Use a deliberately broken resolver: open a non-existent db so we have
	// a nil db field. We test lookup directly via a valid resolver placeholder.
	// Since lookup parses the IP before touching the db, an invalid IP returns
	// early before the db is ever consulted.
	r := &GeoResolver{db: nil}

	got := r.lookup("not-an-ip")
	if got == nil {
		t.Fatal("lookup with invalid IP returned nil, expected &GeoInfo{}")
	}
	if got.Country != "" || got.City != "" {
		t.Errorf("expected zero-value GeoInfo for invalid IP, got %+v", got)
	}
}

// TestGeoResolver_Lookup_EmptyIP verifies that an empty string IP is handled
// the same way as any other unparseable IP.
func TestGeoResolver_Lookup_EmptyIP(t *testing.T) {
	r := &GeoResolver{db: nil}

	got := r.lookup("")
	if got == nil {
		t.Fatal("lookup with empty IP returned nil, expected &GeoInfo{}")
	}
}

// TestRequestGeo_SetsGeoInContext verifies the RequestGeo middleware stores a
// non-nil *GeoInfo in the Gin context under ReqGeoKey and calls c.Next().
// It uses a GeoResolver with a nil db — lookup returns &GeoInfo{} on parse
// failure — which is enough to exercise the middleware wiring.
func TestRequestGeo_SetsGeoInContext(t *testing.T) {
	r := &GeoResolver{db: nil}

	var handlerRan bool
	router := gin.New()
	// Install RequestIP first so GetRequestIP has a value.
	router.Use(RequestIP())
	router.GET("/test", RequestGeo(r), func(c *gin.Context) {
		handlerRan = true

		v, exists := c.Get(ReqGeoKey)
		if !exists {
			t.Error("expected ReqGeoKey to be set in context")
			c.Status(http.StatusInternalServerError)
			return
		}
		info, ok := v.(*GeoInfo)
		if !ok {
			t.Errorf("expected *GeoInfo at ReqGeoKey, got %T", v)
			c.Status(http.StatusInternalServerError)
			return
		}
		if info == nil {
			t.Error("expected non-nil *GeoInfo in context")
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

	if !handlerRan {
		t.Error("downstream handler was never called — c.Next() may not have been invoked")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

// TestGeoInfo_ZeroValue documents that the zero value of GeoInfo is a valid,
// safely usable struct (all fields are zero/empty/false).
func TestGeoInfo_ZeroValue(t *testing.T) {
	var info GeoInfo
	if info.Country != "" {
		t.Errorf("expected empty Country, got %q", info.Country)
	}
	if info.IsEU {
		t.Error("expected IsEU to be false for zero value")
	}
	if info.Latitude != 0 || info.Longitude != 0 {
		t.Errorf("expected zero coordinates, got lat=%f lon=%f", info.Latitude, info.Longitude)
	}
}
