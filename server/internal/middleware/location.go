package middleware

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

const ReqGeoKey = "reqGeo"

type GeoInfo struct {
	Country     string // ISO code, e.g. "US"
	CountryName string
	Region      string // first subdivision name
	RegionCode  string // first subdivision ISO code
	City        string
	PostalCode  string
	Timezone    string // IANA, e.g. "America/New_York"
	Latitude    float64
	Longitude   float64
	AccuracyKm  uint16
	Continent   string // code, e.g. "NA"
	IsEU        bool
}

type GeoResolver struct {
	db *geoip2.Reader
}

// NewGeoResolver opens the mmdb once at startup. Close it on shutdown.
func NewGeoResolver(dbPath string) (*GeoResolver, error) {
	db, err := geoip2.Open(dbPath)
	fmt.Println("aaa")
	fmt.Println("bbb")
	if err != nil {
		return nil, err
	}
	return &GeoResolver{db: db}, nil
}

func (g *GeoResolver) Close() error { return g.db.Close() }

// RequestGeo must run after RequestIP.
func RequestGeo(g *GeoResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ReqGeoKey, g.lookup(GetRequestIP(c)))
		c.Next()
	}
}

func (g *GeoResolver) lookup(rawIP string) *GeoInfo {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return &GeoInfo{}
	}

	rec, err := g.db.City(ip)
	if err != nil {
		return &GeoInfo{}
	}

	info := &GeoInfo{
		Country:     rec.Country.IsoCode,
		CountryName: rec.Country.Names["en"],
		City:        rec.City.Names["en"],
		PostalCode:  rec.Postal.Code,
		Timezone:    rec.Location.TimeZone,
		Latitude:    rec.Location.Latitude,
		Longitude:   rec.Location.Longitude,
		AccuracyKm:  rec.Location.AccuracyRadius,
		Continent:   rec.Continent.Code,
		IsEU:        rec.Country.IsInEuropeanUnion,
	}
	if len(rec.Subdivisions) > 0 {
		info.Region = rec.Subdivisions[0].Names["en"]
		info.RegionCode = rec.Subdivisions[0].IsoCode
	}
	return info
}

// GetRequestGeo never returns nil.
func GetRequestGeo(c *gin.Context) *GeoInfo {
	if v, ok := c.Get(ReqGeoKey); ok {
		if info, ok := v.(*GeoInfo); ok {
			return info
		}
	}
	return &GeoInfo{}
}
