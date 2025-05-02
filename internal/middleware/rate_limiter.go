package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Rate limiting configuration
type RateLimiter struct {
	// IP address -> last seen time
	ips map[string][]time.Time
	mu  sync.Mutex
	// Max requests per time window
	max int
	// Time window duration
	window time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		ips:    make(map[string][]time.Time),
		max:    max,
		window: window,
	}
}

// RateLimitMiddleware limits the number of requests from a single IP
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			// If there's an error, use the RemoteAddr directly
			ip = c.Request.RemoteAddr
		}

		// For extra security, also check X-Forwarded-For header
		// if our application is behind a proxy or load balancer
		forwardedFor := c.Request.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			// Use the first IP in the list
			ip = forwardedFor
		}

		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Get current time
		now := time.Now()

		// Initialize if this is the first request from this IP
		if _, exists := rl.ips[ip]; !exists {
			rl.ips[ip] = []time.Time{now}
			c.Next()
			return
		}

		// Filter out timestamps that are outside of our window
		var requests []time.Time
		for _, timestamp := range rl.ips[ip] {
			if now.Sub(timestamp) <= rl.window {
				requests = append(requests, timestamp)
			}
		}

		// Add current request
		requests = append(requests, now)
		rl.ips[ip] = requests

		// Check if we've exceeded our limit
		if len(requests) > rl.max {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupTask starts a background goroutine to clean up old IP records
func (rl *RateLimiter) CleanupTask() {
	go func() {
		for {
			time.Sleep(time.Minute)

			rl.mu.Lock()
			now := time.Now()

			// For each IP, filter out old timestamps
			for ip, timestamps := range rl.ips {
				var newTimestamps []time.Time
				for _, timestamp := range timestamps {
					if now.Sub(timestamp) <= rl.window {
						newTimestamps = append(newTimestamps, timestamp)
					}
				}

				// If no timestamps remain, remove the IP entirely
				if len(newTimestamps) == 0 {
					delete(rl.ips, ip)
				} else {
					rl.ips[ip] = newTimestamps
				}
			}

			rl.mu.Unlock()
		}
	}()
}
