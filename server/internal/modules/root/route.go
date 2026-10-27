package root

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, rateLimiter *RateLimiter, handler *Handler) {
	router.GET("/", rateLimiter.RootRateLimiter(), handler.HandleRoot)
}
