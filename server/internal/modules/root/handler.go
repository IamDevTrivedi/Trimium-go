package root

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) HandleRoot(c *gin.Context) {
	data := h.service.GetStatus()
	c.JSON(http.StatusOK, data)
}
