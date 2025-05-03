package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
)

// HealthCheck provides a simple endpoint to verify the API is running
func HealthCheck(c *gin.Context) {
	// Check database connectivity
	sqlDB, err := database.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database connection not available",
			"time":    time.Now().Format(time.RFC3339),
		})
		return
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database ping failed",
			"time":    time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API is healthy",
		"time":    time.Now().Format(time.RFC3339),
	})
}
