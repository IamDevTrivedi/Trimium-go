package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/logger"
)

// RequestLogger returns a Gin middleware that emits one structured JSON log
// entry per HTTP request. Severity scales with the response status:
// 5xx logs at ERROR, 4xx logs at WARN, everything else logs at INFO.
func RequestLogger() gin.HandlerFunc {
	return requestLogger(logger.Get())
}

// requestLogger builds the middleware around a specific logger so tests can
// inject a buffer-backed handler.
func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process the request first so the entry carries the final
		// status, response size and latency.
		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_bytes", c.Request.ContentLength,
			"response_bytes", c.Writer.Size(),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case status >= http.StatusInternalServerError:
			log.Error("http request", attrs...)
		case status >= http.StatusBadRequest:
			log.Warn("http request", attrs...)
		default:
			log.Info("http request", attrs...)
		}
	}
}
