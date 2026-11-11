package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/middleware"
)

const (
	signUpInitRateLimitPerMinute = 10
)

func NewRateLimiter(store *database.Redis) gin.HandlerFunc {
	return middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  signUpInitRateLimitPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:init:" + middleware.GetRequestIP(c)
		},
	})
}
