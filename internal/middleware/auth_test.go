package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
	cfg    *config.Config
}

func (s *AuthMiddlewareTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Fatal("Failed to connect to database:", err)
	}
	s.db = db
	database.DB = db

	// Auto migrate the models
	err = db.AutoMigrate(
		&models.User{},
		&models.BlacklistedToken{},
		&models.RefreshToken{},
	)
	if err != nil {
		s.T().Fatal("Failed to migrate database:", err)
	}

	// Setup JWT configuration
	s.cfg = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test_secret_key_for_authentication_middleware_tests",
			AccessExpiry:  time.Minute * 15,
			RefreshExpiry: time.Hour * 24,
		},
	}
	middleware.SetConfig(s.cfg)

	// Setup router
	router := gin.New()
	s.router = router

	// Protected route
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		// Get the userID from the context
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Protected route accessed successfully",
			"user_id": userID,
		})
	})

	// Public route for comparison
	router.GET("/api/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Public route accessed successfully",
		})
	})
}

func (s *AuthMiddlewareTestSuite) TearDownSuite() {
	// Close the database connection
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	// Clear the database before each test
	s.db.Exec("DELETE FROM users")
	s.db.Exec("DELETE FROM blacklisted_tokens")
	s.db.Exec("DELETE FROM refresh_tokens")
}

func (s *AuthMiddlewareTestSuite) TestProtectedRouteWithValidToken() {
	// Create a user
	user := models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}
	s.db.Create(&user)

	// Create JWT token
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.cfg.JWT.Secret))

	// Create a test request with the token
	req, _ := http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

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
	assert.Equal(s.T(), "Protected route accessed successfully", response["message"])
	assert.Equal(s.T(), float64(user.ID), response["user_id"])
}

func (s *AuthMiddlewareTestSuite) TestProtectedRouteWithoutToken() {
	// Create a test request without token
	req, _ := http.NewRequest("GET", "/api/protected", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)

	// Check the response structure
	assert.Equal(s.T(), "error", response["status"])
	assert.Equal(s.T(), "Authorization required", response["error"])
}

func (s *AuthMiddlewareTestSuite) TestProtectedRouteWithInvalidToken() {
	// Create a test request with invalid token
	req, _ := http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *AuthMiddlewareTestSuite) TestProtectedRouteWithExpiredToken() {
	// Create a user
	user := models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}
	s.db.Create(&user)

	// Create JWT token with past expiration
	expirationTime := time.Now().Add(-15 * time.Minute) // Expired
	claims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-30 * time.Minute)),
			Subject:   "1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.cfg.JWT.Secret))

	// Create a test request with the expired token
	req, _ := http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *AuthMiddlewareTestSuite) TestProtectedRouteWithBlacklistedToken() {
	// Create a user
	user := models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}
	s.db.Create(&user)

	// Create JWT token
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.cfg.JWT.Secret))

	// Blacklist the token
	blacklistedToken := models.BlacklistedToken{
		Token:     tokenString,
		ExpiresAt: expirationTime,
	}
	s.db.Create(&blacklistedToken)

	// Create a test request with the blacklisted token
	req, _ := http.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *AuthMiddlewareTestSuite) TestPublicRouteAccess() {
	// Create a test request to public route
	req, _ := http.NewRequest("GET", "/api/public", nil)

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
	assert.Equal(s.T(), "Public route accessed successfully", response["message"])
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
