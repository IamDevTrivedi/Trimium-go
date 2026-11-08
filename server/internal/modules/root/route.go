package root

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, indexRL, welcomeEmailRL gin.HandlerFunc, h Handler) {
	router.GET("/api/v1", indexRL, h.HandleRoot)
	router.POST("/api/v1/emails/welcome", welcomeEmailRL, h.HandleSendWelcomeEmail)
}
