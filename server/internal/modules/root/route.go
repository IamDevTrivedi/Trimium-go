package root

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, indexRL gin.HandlerFunc, h Handler) {
	router.GET("/api/v1", indexRL, h.HandleRoot)
}
