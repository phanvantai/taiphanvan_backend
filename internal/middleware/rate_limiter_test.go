package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("Rate limit not exceeded", func(t *testing.T) {
		// Create a new rate limiter with 3 requests per second
		limiter := middleware.NewRateLimiter(3, time.Second)

		// Setup router with the rate limiter middleware
		router := gin.New()
		router.Use(limiter.RateLimitMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Make 3 requests (should all succeed)
		for i := 0; i < 3; i++ {
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			w = httptest.NewRecorder() // Reset the recorder for next request
		}
	})

	t.Run("Rate limit exceeded", func(t *testing.T) {
		// Create a new rate limiter with 3 requests per second
		limiter := middleware.NewRateLimiter(3, time.Second)

		// Setup router with the rate limiter middleware
		router := gin.New()
		router.Use(limiter.RateLimitMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"

		// Make 3 requests (should all succeed)
		for i := 0; i < 3; i++ {
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			w = httptest.NewRecorder() // Reset the recorder for next request
		}

		// Make a 4th request (should be rate limited)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Different IP addresses", func(t *testing.T) {
		// Create a new rate limiter with 2 requests per second
		limiter := middleware.NewRateLimiter(2, time.Second)

		// Setup router with the rate limiter middleware
		router := gin.New()
		router.Use(limiter.RateLimitMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create test requests from different IPs
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.3:12345"

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.4:12345"

		// Make 2 requests from IP1 (should succeed)
		for i := 0; i < 2; i++ {
			router.ServeHTTP(w1, req1)
			assert.Equal(t, http.StatusOK, w1.Code)
			w1 = httptest.NewRecorder()
		}

		// Make 2 requests from IP2 (should succeed)
		for i := 0; i < 2; i++ {
			router.ServeHTTP(w2, req2)
			assert.Equal(t, http.StatusOK, w2.Code)
			w2 = httptest.NewRecorder()
		}

		// Make one more request from each IP (should be rate limited)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusTooManyRequests, w1.Code)

		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})

	t.Run("X-Forwarded-For header", func(t *testing.T) {
		// Create a new rate limiter with 2 requests per second
		limiter := middleware.NewRateLimiter(2, time.Second)

		// Setup router with the rate limiter middleware
		router := gin.New()
		router.Use(limiter.RateLimitMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create a test request with X-Forwarded-For header
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345" // This will be ignored in favor of X-Forwarded-For
		req.Header.Set("X-Forwarded-For", "192.168.1.5")

		// Make 2 requests (should succeed)
		for i := 0; i < 2; i++ {
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			w = httptest.NewRecorder()
		}

		// Make a 3rd request (should be rate limited)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Rate limit window expires", func(t *testing.T) {
		// Create a new rate limiter with 1 request per 100ms
		limiter := middleware.NewRateLimiter(1, 100*time.Millisecond)

		// Setup router with the rate limiter middleware
		router := gin.New()
		router.Use(limiter.RateLimitMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"

		// First request should succeed
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		w = httptest.NewRecorder()

		// Second request immediately after should be rate limited
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		w = httptest.NewRecorder()

		// Wait for the window to expire
		time.Sleep(110 * time.Millisecond)

		// Now the request should succeed again
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
