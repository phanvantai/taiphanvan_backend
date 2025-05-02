package logger

import (
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/personal_blog_backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// Setup initializes the logger with the appropriate configuration
func Setup(cfg *config.Config) {
	// Determine output format (console or JSON)
	var output io.Writer = os.Stdout

	// Configure output format
	if cfg.Logging.Format == "json" {
		// Use JSON format for production (easier to parse by log aggregation tools)
		Logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// Use pretty console format for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		Logger = zerolog.New(output).With().Timestamp().Logger()
	}

	// Set log level
	switch cfg.Logging.Level {
	case "debug":
		Logger = Logger.Level(zerolog.DebugLevel)
	case "info":
		Logger = Logger.Level(zerolog.InfoLevel)
	case "warn":
		Logger = Logger.Level(zerolog.WarnLevel)
	case "error":
		Logger = Logger.Level(zerolog.ErrorLevel)
	default:
		Logger = Logger.Level(zerolog.InfoLevel)
	}

	// Set as global logger
	log.Logger = Logger

	Logger.Info().
		Str("level", cfg.Logging.Level).
		Str("format", cfg.Logging.Format).
		Msg("Logger initialized")
}

// GinMiddleware returns a Gin middleware for structured logging
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get request latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get method
		method := c.Request.Method

		// Get status code
		statusCode := c.Writer.Status()

		// Get error messages if any
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Add full path if there are query parameters
		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request details
		event := Logger.Info()

		// Change log level based on status code
		if statusCode >= 500 {
			event = Logger.Error()
		} else if statusCode >= 400 {
			event = Logger.Warn()
		}

		// Add structured context fields
		event.
			Int("status", statusCode).
			Str("method", method).
			Str("path", path).
			Str("ip", clientIP).
			Dur("latency", latency).
			Str("latency_human", latency.String())

		// Add error message if present
		if errorMessage != "" {
			event.Str("error", errorMessage)
		}

		// Include user ID if authenticated
		if userID, exists := c.Get("userID"); exists {
			event.Interface("user_id", userID)
		}

		// Include request ID if present
		if requestID := c.Writer.Header().Get("X-Request-ID"); requestID != "" {
			event.Str("request_id", requestID)
		}

		event.Msg("Request processed")
	}
}
