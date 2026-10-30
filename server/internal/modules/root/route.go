package root

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, rl RateLimiter, h Handler) {
	router.GET("/api/v1", rl.Index, h.HandleRoot)
}
