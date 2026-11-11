package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

type Handler interface {
	HandleSignUpInit(c *gin.Context)
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
