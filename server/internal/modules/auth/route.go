package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, signUpInitRL gin.HandlerFunc, h Handler) {
	router.POST("/api/v1/auth/signup/init", signUpInitRL, h.HandleSignUpInit)
}
