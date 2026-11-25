package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/middleware"
)

const (
	signUpInitRateLimitPerMinute          = 5
	signUpVerifyRateLimitPerMinute        = 10
	createUserRateLimitPerMinute          = 5
	resetPasswordInitRateLimitPerMinute   = 5
	resetPasswordVerifyRateLimitPerMinute = 10
	resetPasswordRateLimitPerMinute       = 5
)

func NewRateLimiters(store *database.Redis) (
	signUpInitRL, signUpVerifyRL, signUpCreateRL, resetPasswordInitRL, resetPasswordVerifyRL, resetPasswordRL gin.HandlerFunc,
) {
	signUpInitRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  signUpInitRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:init:" + middleware.GetRequestIP(c)
		},
	})

	signUpVerifyRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  signUpVerifyRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:verify:" + middleware.GetRequestIP(c)
		},
	})

	signUpCreateRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  createUserRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:signup:create" + middleware.GetRequestIP(c)
		},
	})

	resetPasswordInitRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  resetPasswordInitRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:reset-password:init:" + middleware.GetRequestIP(c)
		},
	})

	resetPasswordVerifyRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  resetPasswordVerifyRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:reset-password:verify:" + middleware.GetRequestIP(c)
		},
	})

	resetPasswordRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  resetPasswordRateLimitPerMinute,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "auth:reset-password:" + middleware.GetRequestIP(c)
		},
	})

	return signUpInitRL, signUpVerifyRL, signUpCreateRL, resetPasswordInitRL, resetPasswordVerifyRL, resetPasswordRL
}
