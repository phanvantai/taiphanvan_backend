package handlers

import (
	"bytes"
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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of the database
type MockDB struct {
	mock.Mock
}

// AuthTestSuite is a test suite for auth handlers
type AuthTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
}

// SetupSuite sets up the test suite
func (s *AuthTestSuite) SetupSuite() {
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
	jwtConfig := config.JWTConfig{
		Secret:        "test_secret_key_for_testing_purposes_only",
		AccessExpiry:  time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
	}
	appConfig := &config.Config{
		JWT: jwtConfig,
	}
	middleware.SetConfig(appConfig)

	// Setup router
	router := gin.Default()
	s.router = router

	// Register routes
	router.POST("/api/auth/register", Register)
	router.POST("/api/auth/login", Login)
	router.POST("/api/auth/refresh", RefreshToken)
	router.POST("/api/auth/revoke", RevokeToken)

	// Protected routes
	authGroup := router.Group("/api")
	authGroup.Use(func(c *gin.Context) {
		// Mock middleware for testing
		userID, exists := c.Get("userID")
		if !exists {
			// For testing, set a default user ID
			c.Set("userID", uint(1))
		} else {
			c.Set("userID", userID)
		}
		c.Next()
	})

	authGroup.GET("/profile", GetProfile)
	authGroup.PUT("/profile", UpdateProfile)
}

// TearDownSuite tears down the test suite
func (s *AuthTestSuite) TearDownSuite() {
	// Close the database connection
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest sets up each test
func (s *AuthTestSuite) SetupTest() {
	// Clear the database before each test
	s.db.Exec("DELETE FROM users")
	s.db.Exec("DELETE FROM refresh_tokens")
	s.db.Exec("DELETE FROM blacklisted_tokens")
}

// TestRegisterWithValidInput tests the Register handler with valid input
func (s *AuthTestSuite) TestRegisterWithValidInput() {
	// Create a test request
	registerRequest := models.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	jsonData, _ := json.Marshal(registerRequest)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)

	// Check the response structure
	assert.Equal(s.T(), "success", response["status"])
	assert.Equal(s.T(), "User registered successfully", response["message"])

	// Check that the user was created in the database
	var user models.User
	result := s.db.Where("email = ?", registerRequest.Email).First(&user)
	assert.NoError(s.T(), result.Error)
	assert.Equal(s.T(), registerRequest.Username, user.Username)
	assert.Equal(s.T(), registerRequest.Email, user.Email)
	assert.Equal(s.T(), registerRequest.FirstName, user.FirstName)
	assert.Equal(s.T(), registerRequest.LastName, user.LastName)
	assert.Equal(s.T(), "user", user.Role) // Default role
}

// TestRegisterWithExistingEmail tests the Register handler with an existing email
func (s *AuthTestSuite) TestRegisterWithExistingEmail() {
	// Create a user first
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := models.User{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  string(hashedPassword),
		FirstName: "Existing",
		LastName:  "User",
		Role:      "user",
	}
	s.db.Create(&existingUser)

	// Try to register with the same email
	registerRequest := models.RegisterRequest{
		Username:  "newuser",
		Email:     "existing@example.com", // Same email
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	jsonData, _ := json.Marshal(registerRequest)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	s.router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(s.T(), http.StatusConflict, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)

	// Check the response structure
	assert.Equal(s.T(), "error", response["status"])
	assert.Equal(s.T(), "User exists", response["error"])
}

// TestLoginWithValidCredentials tests the Login handler with valid credentials
func (s *AuthTestSuite) TestLoginWithValidCredentials() {
	// Create a user first
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	s.db.Create(&user)

	// Login with valid credentials
	loginRequest := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	jsonData, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(s.T(), "User logged in successfully", response["message"])

	// Check that the response contains token data
	data, ok := response["data"].(map[string]interface{})
	assert.True(s.T(), ok)
	assert.NotEmpty(s.T(), data["access_token"])
	assert.NotEmpty(s.T(), data["refresh_token"])
	assert.Equal(s.T(), "Bearer", data["token_type"])
	assert.NotZero(s.T(), data["expires_in"])
}

// TestLoginWithInvalidCredentials tests the Login handler with invalid credentials
func (s *AuthTestSuite) TestLoginWithInvalidCredentials() {
	// Create a user first
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	s.db.Create(&user)

	// Login with invalid password
	loginRequest := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	jsonData, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(s.T(), "Authentication failed", response["error"])
}

// TestRefreshTokenSuccessfully tests the RefreshToken handler
func (s *AuthTestSuite) TestRefreshTokenSuccessfully() {
	// Create a user first
	user := models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword", // Not used in this test
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	s.db.Create(&user)

	// Create a refresh token
	expirationTime := time.Now().Add(time.Hour * 24)
	claims := &middleware.Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshTokenString, _ := token.SignedString([]byte(middleware.AppConfig.JWT.Secret))

	// Store the refresh token in the database
	refreshToken := models.RefreshToken{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: expirationTime,
		IssuedAt:  time.Now(),
		Revoked:   false,
	}
	s.db.Create(&refreshToken)

	// Request to refresh the token
	refreshRequest := models.RefreshTokenRequest{
		RefreshToken: refreshTokenString,
	}

	jsonData, _ := json.Marshal(refreshRequest)
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(s.T(), "Access token refreshed successfully", response["message"])

	// Check that the response contains a new access token
	data, ok := response["data"].(map[string]interface{})
	assert.True(s.T(), ok)
	assert.NotEmpty(s.T(), data["access_token"])
	assert.Equal(s.T(), "Bearer", data["token_type"])
	assert.NotZero(s.T(), data["expires_in"])
}

// TestRevokeTokenSuccessfully tests the RevokeToken handler
func (s *AuthTestSuite) TestRevokeTokenSuccessfully() {
	// Create a user first
	user := models.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword", // Not used in this test
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	s.db.Create(&user)

	// Create a refresh token
	refreshTokenString := "valid_refresh_token"
	refreshToken := models.RefreshToken{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24),
		IssuedAt:  time.Now(),
		Revoked:   false,
	}
	s.db.Create(&refreshToken)

	// Request to revoke the token
	revokeRequest := models.TokenRevokeRequest{
		RefreshToken: refreshTokenString,
	}

	jsonData, _ := json.Marshal(revokeRequest)
	req, _ := http.NewRequest("POST", "/api/auth/revoke", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(s.T(), "Token revoked successfully", response["message"])

	// Check that the token was revoked in the database
	var updatedToken models.RefreshToken
	s.db.Where("token = ?", refreshTokenString).First(&updatedToken)
	assert.True(s.T(), updatedToken.Revoked)
}

// TestGetProfileSuccessfully tests the GetProfile handler
func (s *AuthTestSuite) TestGetProfileSuccessfully() {
	// Create a user first
	user := models.User{
		ID:        1, // This ID matches the one set in the mock middleware
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Bio:       "Test bio",
		Role:      "user",
	}
	s.db.Create(&user)

	// Create a request
	req, _ := http.NewRequest("GET", "/api/profile", nil)

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
	assert.Equal(s.T(), "User profile retrieved successfully", response["message"])

	// Check the user data
	data, ok := response["data"].(map[string]interface{})
	assert.True(s.T(), ok)
	assert.Equal(s.T(), float64(user.ID), data["id"])
	assert.Equal(s.T(), user.Username, data["username"])
	assert.Equal(s.T(), user.Email, data["email"])
	assert.Equal(s.T(), user.FirstName, data["first_name"])
	assert.Equal(s.T(), user.LastName, data["last_name"])
	assert.Equal(s.T(), user.Bio, data["bio"])
}

// TestUpdateProfileSuccessfully tests the UpdateProfile handler
func (s *AuthTestSuite) TestUpdateProfileSuccessfully() {
	// Create a user first
	user := models.User{
		ID:        1, // This ID matches the one set in the mock middleware
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Bio:       "Original bio",
		Role:      "user",
	}
	s.db.Create(&user)

	// Update request
	updateRequest := struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Bio          string `json:"bio"`
		ProfileImage string `json:"profile_image"`
	}{
		FirstName:    "Updated",
		LastName:     "Name",
		Bio:          "Updated bio",
		ProfileImage: "https://example.com/image.jpg",
	}

	jsonData, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/api/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(s.T(), "Profile updated successfully", response["message"])

	// Check that the user was updated in the database
	var updatedUser models.User
	s.db.First(&updatedUser, user.ID)
	assert.Equal(s.T(), updateRequest.FirstName, updatedUser.FirstName)
	assert.Equal(s.T(), updateRequest.LastName, updatedUser.LastName)
	assert.Equal(s.T(), updateRequest.Bio, updatedUser.Bio)
	assert.Equal(s.T(), updateRequest.ProfileImage, updatedUser.ProfileImage)
}

// TestAuthSuite runs the test suite
func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
