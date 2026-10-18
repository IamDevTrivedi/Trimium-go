package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response is the uniform JSON envelope returned by every API endpoint.
type Response struct {
	Success   bool   `json:"success"`
	Timestamp int64  `json:"timestamp"`
	Msg       string `json:"msg"`
	Metadata  any    `json:"metadata"`
}

// NewResponse builds the envelope with the current time in epoch milliseconds.
func NewResponse(success bool, msg string, metadata any) Response {
	return Response{
		Success:   success,
		Timestamp: time.Now().UnixMilli(),
		Msg:       msg,
		Metadata:  metadata,
	}
}

// JSON writes the uniform envelope with full control over every field.
func JSON(c *gin.Context, statusCode int, success bool, msg string, metadata any) {
	c.JSON(statusCode, NewResponse(success, msg, metadata))
}

// Success writes a 200 OK success envelope.
func Success(c *gin.Context, msg string, metadata any) {
	JSON(c, http.StatusOK, true, msg, metadata)
}

// Created writes a 201 Created success envelope.
func Created(c *gin.Context, msg string, metadata any) {
	JSON(c, http.StatusCreated, true, msg, metadata)
}

// Error writes an error envelope with nil metadata.
func Error(c *gin.Context, statusCode int, msg string) {
	JSON(c, statusCode, false, msg, nil)
}
