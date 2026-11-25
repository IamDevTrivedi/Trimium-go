package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router gin.IRouter,
	signUpInitRL, signUpVerifyRL, signUpCreateRL, resetPasswordInitRL, resetPasswordVerifyRL, resetPasswordRL gin.HandlerFunc,
	h Handler,
) {
	router.POST("/api/v1/auth/signup/init", signUpInitRL, h.HandleSignUpInit)
	router.POST("/api/v1/auth/signup/verify", signUpVerifyRL, h.HandleSignUpVerify)
	router.POST("/api/v1/auth/signup/create", signUpCreateRL, h.HandleCreateUser)
	router.POST("/api/v1/auth/reset-password/init", resetPasswordInitRL, h.HandleResetPasswordInit)
	router.POST("/api/v1/auth/reset-password/verify", resetPasswordVerifyRL, h.HandleResetPasswordVerify)
	router.POST("/api/v1/auth/reset-password/reset", resetPasswordRL, h.HandleResetPassword)
}
