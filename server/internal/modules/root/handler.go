package root

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

type Handler interface {
	HandleRoot(c *gin.Context)
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

func (h *handler) HandleRoot(c *gin.Context) {
	data, err := h.service.GetAppInfo(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "App is not working")
		return
	}

	response.Success(c, "Welcome to Trimium API", data)
}
