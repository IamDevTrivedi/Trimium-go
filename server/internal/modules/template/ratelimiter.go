package template

import (
	"time"

	"server/internal/database"
	"server/internal/middleware"

	"github.com/gin-gonic/gin"
)

const (
	indexRateLimitPerMinute = 50
)

func NewRateLimiter(store *database.Redis) (indexRL gin.HandlerFunc) {
	indexRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  indexRateLimitPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "template:index:" + middleware.GetRequestIP(c)
		},
	})

	return indexRL
}
