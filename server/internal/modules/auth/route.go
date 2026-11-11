package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, signUpInitRL, signUpVerifyRL gin.HandlerFunc, h Handler) {
	router.POST("/api/v1/auth/signup/init", signUpInitRL, h.HandleSignUpInit)
	router.POST("/api/v1/auth/signup/verify", signUpVerifyRL, h.HandleSignUpVerify)
}
