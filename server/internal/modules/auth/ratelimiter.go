package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/middleware"
)

const (
	signUpInitRateLimitPerMinute   = 10
	signUpVerifyRateLimitPerMinute = 10
)

func NewRateLimiters(store *database.Redis) (signUpInitRL, signUpVerifyRL gin.HandlerFunc) {
	signUpInitRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  signUpInitRateLimitPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:init:" + middleware.GetRequestIP(c)
		},
	})

	signUpVerifyRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  signUpVerifyRateLimitPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:verify:" + middleware.GetRequestIP(c)
		},
	})

	return signUpInitRL, signUpVerifyRL
}
