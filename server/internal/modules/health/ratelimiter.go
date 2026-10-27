package health

import (
	"time"

	"server/internal/database"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	limiter gin.HandlerFunc
}

func NewRateLimiter(redisStore *database.Redis) *RateLimiter {
	return &RateLimiter{
		limiter: middleware.RateLimiter(middleware.RateLimiterConfig{
			Client: redisStore.Client(),
			Limit:  5,
			Window: time.Minute,
			KeyGen: func(c *gin.Context) string {
				return middleware.GetRequestIP(c)
			},
		}),
	}
}

func (r *RateLimiter) HealthRateLimiter(c *gin.Context) {
	r.limiter(c)
}
