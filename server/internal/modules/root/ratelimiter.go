package root

import (
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/middleware"
)

const rateLimitPerMinute = 50

type RateLimiter interface {
	Index(c *gin.Context)
}

type rateLimiter struct {
	index gin.HandlerFunc
}

func NewRateLimiter(store *database.Redis) RateLimiter {
	return &rateLimiter{
		index: middleware.RateLimiter(middleware.RateLimiterConfig{
			Client: store.Client(),
			Limit:  rateLimitPerMinute,
			Window: 1 * time.Minute,
			KeyGen: func(c *gin.Context) string {
				return "root:index:" + middleware.GetRequestIP(c)
			},
		}),
	}
}

func (r *rateLimiter) Index(c *gin.Context) { r.index(c) }
