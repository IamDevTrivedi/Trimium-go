package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

const ReqUAKey = "reqUA"

type UserAgentInfo struct {
	Raw            string
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	Device         string
	DeviceType     string
	IsBot          bool
}

func RequestUserAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ReqUAKey, parseUA(c.GetHeader("User-Agent")))
		c.Next()
	}
}

func parseUA(raw string) *UserAgentInfo {
	ua := useragent.Parse(raw)

	return &UserAgentInfo{
		Raw:            raw,
		Browser:        ua.Name,
		BrowserVersion: ua.Version,
		OS:             ua.OS,
		OSVersion:      ua.OSVersion,
		Device:         ua.Device,
		DeviceType:     deviceType(ua),
		IsBot:          ua.Bot,
	}
}

func deviceType(ua useragent.UserAgent) string {
	switch {
	case ua.Bot:
		return "bot"
	case ua.Tablet:
		return "tablet"
	case ua.Mobile:
		return "mobile"
	case ua.Desktop:
		return "desktop"
	default:
		return ""
	}
}

// GetRequestUA never returns nil.
func GetRequestUA(c *gin.Context) *UserAgentInfo {
	if v, ok := c.Get(ReqUAKey); ok {
		if info, ok := v.(*UserAgentInfo); ok {
			return info
		}
	}
	return &UserAgentInfo{}
}
