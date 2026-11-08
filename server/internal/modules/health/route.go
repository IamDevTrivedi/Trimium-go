package health

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, indexRL, redisRL, mongoRL gin.HandlerFunc, h Handler) {
	router.GET("/api/v1/healthz", indexRL, h.HandleHealth)
	router.GET("/api/v1/healthz/redis", redisRL, h.HandleRedisHealth)
	router.GET("/api/v1/healthz/mongo", mongoRL, h.HandleMongoHealth)
}
