package root

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the root module endpoints into the router.
func RegisterRoutes(router *gin.Engine, handler *Handler) {
	router.GET("/", handler.HandleRoot)
}
