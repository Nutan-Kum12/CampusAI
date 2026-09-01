package middleware

import (
	"log/slog"
	"time"

	"github.com/Nutan-Kum12/CampusAI/pkg/response"
	"github.com/gin-gonic/gin"
)

//	2xx → INFO   (normal)
//	4xx → WARN   (client made a mistake)
//	5xx → ERROR  (our mistake)
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // run handler

		latency := time.Since(start)
		status := c.Writer.Status()

		logFn := slog.Info
		if status >= 500 {
			logFn = slog.Error
		} else if status >= 400 {
			logFn = slog.Warn
		}

		logFn("HTTP",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("Panic recovered",
					"error", recovered,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
				response.InternalError(c, "An unexpected error occurred")
			}
		}()
		c.Next()
	}
}
