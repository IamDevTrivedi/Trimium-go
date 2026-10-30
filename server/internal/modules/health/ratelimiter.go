package health

import (
	"time"

	"server/internal/database"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Index(c *gin.Context)
	Redis(c *gin.Context)
	Mongo(c *gin.Context)
}

type rateLimiter struct {
	index gin.HandlerFunc
	redis gin.HandlerFunc
	mongo gin.HandlerFunc
}

func NewRateLimiter(store *database.Redis) RateLimiter {
	return &rateLimiter{
		index: middleware.RateLimiter(middleware.RateLimiterConfig{
			Client: store.Client(),
			Limit:  5,
			Window: 1 * time.Minute,
			KeyGen: func(c *gin.Context) string {
				return "health:index:" + middleware.GetRequestIP(c)
			},
		}),

		redis: middleware.RateLimiter(middleware.RateLimiterConfig{
			Client: store.Client(),
			Limit:  10,
			Window: 1 * time.Minute,
			KeyGen: func(c *gin.Context) string {
				return "health:redis:" + middleware.GetRequestIP(c)
			},
		}),

		mongo: middleware.RateLimiter(middleware.RateLimiterConfig{
			Client: store.Client(),
			Limit:  15,
			Window: 1 * time.Minute,
			KeyGen: func(c *gin.Context) string {
				return "health:mongo:" + middleware.GetRequestIP(c)
			},
		}),
	}
}

func (r *rateLimiter) Index(c *gin.Context) { r.index(c) }
func (r *rateLimiter) Redis(c *gin.Context) { r.redis(c) }
func (r *rateLimiter) Mongo(c *gin.Context) { r.mongo(c) }
