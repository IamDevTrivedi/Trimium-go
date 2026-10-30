package health

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, rl RateLimiter, h Handler) {
	router.GET("/api/v1/healthz", rl.Index, h.HandleHealth)
	router.GET("/api/v1/healthz/redis", rl.Redis, h.HandleRedisHealth)
	router.GET("/api/v1/healthz/mongo", rl.Mongo, h.HandleMongoHealth)
}
