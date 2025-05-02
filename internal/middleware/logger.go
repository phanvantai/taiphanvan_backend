package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs request information
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		// Format log message with color coding by status code
		var statusColor, resetColor string
		if gin.Mode() != gin.ReleaseMode {
			resetColor = "\033[0m"
			if statusCode >= 500 {
				statusColor = "\033[31m" // Red
			} else if statusCode >= 400 {
				statusColor = "\033[33m" // Yellow
			} else if statusCode >= 300 {
				statusColor = "\033[36m" // Cyan
			} else {
				statusColor = "\033[32m" // Green
			}
		}

		if errorMessage != "" {
			log.Printf("%s%d%s | %13v | %15s | %-7s %s | %s",
				statusColor, statusCode, resetColor,
				latency, clientIP, method, path, errorMessage)
		} else {
			log.Printf("%s%d%s | %13v | %15s | %-7s %s",
				statusColor, statusCode, resetColor,
				latency, clientIP, method, path)
		}
	}
}
