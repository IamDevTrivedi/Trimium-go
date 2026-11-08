package health

import (
	"time"

	"server/internal/database"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

const (
	indexRatePerMinute = 5
	redisRatePerMinute = 10
	mongoRatePerMinute = 15
)

func NewRateLimiters(store *database.Redis) (indexRL, redisRL, mongoRL gin.HandlerFunc) {
	indexRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  indexRatePerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "health:index:" + middleware.GetRequestIP(c)
		},
	})

	redisRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  redisRatePerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "health:redis:" + middleware.GetRequestIP(c)
		},
	})

	mongoRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  mongoRatePerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "health:mongo:" + middleware.GetRequestIP(c)
		},
	})

	return indexRL, redisRL, mongoRL
}
