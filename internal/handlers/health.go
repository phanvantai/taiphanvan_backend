package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
)

// HealthCheck godoc
// @Summary Check API health
// @Description Provides a simple endpoint to verify the API and database are running
// @Tags System
// @Produce json
// @Success 200 {object} models.SwaggerStandardResponse "API is healthy"
// @Failure 503 {object} models.SwaggerStandardResponse "Database connection issues"
// @Router /health [get]
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
