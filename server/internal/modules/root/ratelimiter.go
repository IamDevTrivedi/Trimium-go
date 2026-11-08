package root

import (
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/middleware"
)

const (
	rateLimitPerMinute    = 50
	welcomeEmailPerMinute = 5
)

func NewRateLimiters(store *database.Redis) (indexRL, welcomeEmailRL gin.HandlerFunc) {
	indexRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  rateLimitPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "root:index:" + middleware.GetRequestIP(c)
		},
	})

	welcomeEmailRL = middleware.RateLimiter(middleware.RateLimiterConfig{
		Client: store.Client(),
		Limit:  welcomeEmailPerMinute,
		Window: 1 * time.Minute,
		KeyGen: func(c *gin.Context) string {
			return "root:welcome_email:" + middleware.GetRequestIP(c)
		},
	})

	return indexRL, welcomeEmailRL
}
