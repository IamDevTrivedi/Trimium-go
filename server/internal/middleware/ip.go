package middleware

import (
	"net"

	"github.com/gin-gonic/gin"
)

const ReqIPKey = "reqIP"

// MaxMind test IP that resolves to London in GeoLite2
const localhostFallbackIP = "81.2.69.142"

func RequestIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ReqIPKey, resolveIP(c.ClientIP()))
		c.Next()
	}
}

func resolveIP(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return localhostFallbackIP
	}
	return ip.String()
}

func GetRequestIP(c *gin.Context) string {
	return c.GetString(ReqIPKey)
}
