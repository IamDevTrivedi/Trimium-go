package health

import (
	"github.com/gin-gonic/gin"

	"server/internal/response"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) HandleHealth(c *gin.Context) {
	data, err := h.service.GetStatus()

	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Trimium API is healthy", data)
}
