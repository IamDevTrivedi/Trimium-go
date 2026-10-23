package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ReqIDKey = "reqID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ReqIDKey, newRequestID())
		c.Next()
	}
}

func newRequestID() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}
	return fallbackID()
}

// fallbackID: fallback-<epoch ms>-<8 hex chars>
func fallbackID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failed too; use nanosecond time as the entropy source
		n := time.Now().UnixNano()
		b = []byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)}
	}
	return fmt.Sprintf("fallback-%d-%s", time.Now().UnixMilli(), hex.EncodeToString(b))
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(ReqIDKey)
}
