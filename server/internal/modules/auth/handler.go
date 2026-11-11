package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

type Handler interface {
	HandleSignUpInit(c *gin.Context)
	HandleSignUpVerify(c *gin.Context)
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

func (h *handler) HandleSignUpInit(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid email address")
		return
	}

	if err := h.service.SignUpInit(req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to initiate signup")
		return
	}

	response.Success(c, "Signup initiated successfully", nil)
}

func (h *handler) HandleSignUpVerify(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if err := h.service.SignUpVerify(req.Email, req.OTP); err != nil {
		switch {
		case errors.Is(err, ErrOTPNotFound):
			response.Error(c, http.StatusBadRequest, "OTP not found or expired")
		case errors.Is(err, ErrInvalidOTP):
			response.Error(c, http.StatusBadRequest, "Invalid OTP")
		case errors.Is(err, ErrMaxAttemptsExceeded):
			response.Error(c, http.StatusBadRequest, "Max verification attempts exceeded")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to verify signup")
		}
		return
	}

	response.Success(c, "Signup verified successfully", nil)
}
