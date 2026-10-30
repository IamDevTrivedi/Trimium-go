package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

type Handler interface {
	HandleHealth(c *gin.Context)
	HandleRedisHealth(c *gin.Context)
	HandleMongoHealth(c *gin.Context)
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

func (h *handler) HandleHealth(c *gin.Context) {
	data, err := h.service.GetAppStatus(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "App is not working")
		return
	}
	response.Success(c, "App is working", data)
}

func (h *handler) HandleRedisHealth(c *gin.Context) {
	data, err := h.service.GetRedisStatus(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "Redis is not working")
		return
	}
	response.Success(c, "Redis is working", data)
}

func (h *handler) HandleMongoHealth(c *gin.Context) {
	data, err := h.service.GetMongoStatus(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "Mongo is not working")
		return
	}
	response.Success(c, "Mongo is working", data)
}
