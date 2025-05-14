package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a router with the logger middleware
	router := gin.New()
	router.Use(middleware.LoggerMiddleware())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Add a route that returns an error
	router.GET("/error", func(c *gin.Context) {
		c.AbortWithError(http.StatusInternalServerError, gin.Error{
			Err:  gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "Test Error"},
			Type: gin.ErrorTypePrivate,
		})
	})

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	t.Run("Logs successful request", func(t *testing.T) {
		// Create a test request
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Create a response recorder
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusOK, w.Code)

		// Check the log output
		logOutput := buf.String()
		assert.Contains(t, logOutput, "200")
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/test")

		buf.Reset()
	})

	t.Run("Logs request with query parameters", func(t *testing.T) {
		// Create a test request with query parameters
		req, _ := http.NewRequest("GET", "/test?param=value", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Create a response recorder
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusOK, w.Code)

		// Check the log output
		logOutput := buf.String()
		assert.Contains(t, logOutput, "/test?param=value")

		buf.Reset()
	})

	t.Run("Logs error request", func(t *testing.T) {
		// Create a test request that will produce an error
		req, _ := http.NewRequest("GET", "/error", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Create a response recorder
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Check the log output
		logOutput := buf.String()
		assert.Contains(t, logOutput, "500")
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/error")
		// In test mode, error details might be stripped, so we don't check for the exact error message

		buf.Reset()
	})

	t.Run("Different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE"}

		for _, method := range methods {
			// Create a test request
			req, _ := http.NewRequest(method, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Check the log output
			logOutput := buf.String()
			assert.True(t, strings.Contains(logOutput, method), "Log should contain HTTP method: %s", method)

			buf.Reset()
		}
	})
}
