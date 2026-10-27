package root

import (
	"time"

	"server/internal/database"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	RedisStore *database.Redis
}

func NewRateLimiter(redisStore *database.Redis) *RateLimiter {
	return &RateLimiter{RedisStore: redisStore}
}

func (r *RateLimiter) RootRateLimiter() gin.HandlerFunc {
	return middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: r.RedisStore.Client(),
		Limit:  5,
		Window: time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "root:" + middleware.GetRequestIP(c)
		},
	})
}
