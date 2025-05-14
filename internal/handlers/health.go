package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
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
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse("Service Unavailable", "Database connection not available"))
		return
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.NewErrorResponse("Service Unavailable", "Database ping failed"))
		return
	}

	healthData := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(healthData, "API is healthy"))
}
