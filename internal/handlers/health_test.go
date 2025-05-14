package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type HealthCheckTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
}

func (s *HealthCheckTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Fatal("Failed to connect to database:", err)
	}
	s.db = db
	database.DB = db

	// Setup router
	router := gin.Default()
	s.router = router

	// Register health check route
	router.GET("/api/health", handlers.HealthCheck)
}

func (s *HealthCheckTestSuite) TearDownSuite() {
	// Close the database connection
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (s *HealthCheckTestSuite) TestHealthCheckSuccessful() {
	// Create a test request
	req, _ := http.NewRequest("GET", "/api/health", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)

	// Check the response structure
	assert.Equal(s.T(), "success", response["status"])
	assert.Equal(s.T(), "API is healthy", response["message"])

	// Check that the data contains a time field
	data, ok := response["data"].(map[string]interface{})
	assert.True(s.T(), ok)
	assert.Contains(s.T(), data, "time")
}

// This test requires a mock database to properly test failure conditions
// For now, we'll leave it as a basic test of the success path
func (s *HealthCheckTestSuite) TestHealthCheckWithDatabaseError() {
	// In a real test, we would mock the database failure
	// For this example, we're just testing the happy path
	s.T().Skip("Requires database mocking to properly test failure conditions")
}

func TestHealthCheckSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckTestSuite))
}
