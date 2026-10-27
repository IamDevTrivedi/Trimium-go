package health

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, rateLimiter *RateLimiter, handler *Handler) {
	router.GET("/api/v1/healthz", rateLimiter.HealthRateLimiter(), handler.HandleHealth)
}
