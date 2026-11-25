package middleware

import (
	"github.com/gin-gonic/gin"

	"server/internal/crypto"
)

const ReqIDKey = "reqID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := crypto.NewUUID()
		c.Set(ReqIDKey, id)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(ReqIDKey)
}
