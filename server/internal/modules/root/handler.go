package root

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/response"
)

// WelcomeEmailRequest is the accepted body for the welcome-email endpoint.
type WelcomeEmailRequest struct {
	To []string `json:"to" binding:"required,min=1,dive,email"`
}

type Handler interface {
	HandleRoot(c *gin.Context)
	HandleSendWelcomeEmail(c *gin.Context)
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

// HandleSendWelcomeEmail accepts a list of recipients and asks the
// service to queue a welcome email for them. A 202 means the payload
// was accepted for delivery, not that the email was already sent.
func (h *handler) HandleSendWelcomeEmail(c *gin.Context) {
	var req WelcomeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SendWelcomeEmail(c.Request.Context(), req.To); err != nil {
		response.Error(c, http.StatusInternalServerError, "Could not queue welcome email")
		return
	}

	response.JSON(c, http.StatusAccepted, true, "Welcome email queued", nil)
}
