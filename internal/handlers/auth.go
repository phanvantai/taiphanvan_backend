package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User Registration Data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "Email or username already exists"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var request models.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": err.Error(),
		})
		return
	}

	// Check if user already exists - use Count instead of First to avoid "record not found" error
	var count int64
	database.DB.Model(&models.User{}).Where("email = ?", request.Email).Or("username = ?", request.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"error":   "User exists",
			"message": "Email or username already exists",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Str("email", request.Email).Msg("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Server error",
			"message": "Failed to process registration",
		})
		return
	}

	// Create new user
	user := models.User{
		Username:  request.Username,
		Email:     request.Email,
		Password:  string(hashedPassword),
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Role:      "user", // Default role
	}

	if result := database.DB.Create(&user); result.Error != nil {
		log.Error().Err(result.Error).Str("email", request.Email).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to create user",
		})
		return
	}

	log.Info().Str("email", user.Email).Uint("id", user.ID).Msg("User registered successfully")
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User registered successfully",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login godoc
// @Summary Login to the application
// @Description Authenticate a user and return JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login Credentials"
// @Success 200 {object} models.TokenResponse "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Authentication failed"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": err.Error(),
		})
		return
	}

	// Find the user by email
	var user models.User
	if result := database.DB.Where("email = ?", request.Email).First(&user); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Info().Str("email", request.Email).Msg("Login attempt with non-existent email")
		} else {
			log.Error().Err(result.Error).Str("email", request.Email).Msg("Database error during login")
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Authentication failed",
			"message": "Invalid credentials",
		})
		return
	}

	// Compare passwords
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		log.Info().Str("email", request.Email).Msg("Login attempt with incorrect password")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Authentication failed",
			"message": "Invalid credentials",
		})
		return
	}

	// Generate token pair
	accessToken, refreshToken, _, err := middleware.GenerateTokenPair(user)
	if err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Token generation failed",
			"message": "Failed to generate authentication tokens",
		})
		return
	}

	// Calculate expiry time in seconds for access token
	expiresIn := int(middleware.AppConfig.JWT.AccessExpiry.Seconds())

	log.Info().Str("email", user.Email).Uint("id", user.ID).Msg("User logged in successfully")
	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	})
}

// RefreshToken godoc
// @Summary Refresh an access token
// @Description Get a new access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body models.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} models.TokenResponse "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var request models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": err.Error(),
		})
		return
	}

	// Get a new access token
	accessToken, err := middleware.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Invalid refresh token",
			"message": err.Error(),
		})
		return
	}

	// Calculate expiry time in seconds
	expiresIn := int(middleware.AppConfig.JWT.AccessExpiry.Seconds())

	log.Info().Msg("Access token refreshed successfully")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": models.TokenResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		},
	})
}

// RevokeToken godoc
// @Summary Revoke a refresh token
// @Description Invalidate a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body models.TokenRevokeRequest true "Refresh Token"
// @Success 200 {object} map[string]interface{} "Token revoked successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input or token revocation failed"
// @Security BearerAuth
// @Router /auth/revoke [post]
func RevokeToken(c *gin.Context) {
	var request models.TokenRevokeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": err.Error(),
		})
		return
	}

	// Revoke the refresh token
	err := middleware.RevokeRefreshToken(request.RefreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to revoke token")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Token revocation failed",
			"message": err.Error(),
		})
		return
	}

	log.Info().Msg("Refresh token revoked successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Token revoked successfully",
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Retrieve the current user's profile information
// @Tags Users
// @Produce json
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Security BearerAuth
// @Router /profile [get]
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if result := database.DB.Select("id, username, email, first_name, last_name, bio, role, profile_image, created_at, updated_at").Where("id = ?", userID).First(&user); result.Error != nil {
		log.Warn().Err(result.Error).Interface("user_id", userID).Msg("User not found when fetching profile")
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"error":   "Not found",
			"message": "User not found",
		})
		return
	}

	log.Info().Interface("user_id", userID).Msg("User profile retrieved")
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Param profile body object true "Profile Data"
// @Success 200 {object} map[string]interface{} "Profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Security BearerAuth
// @Router /profile [put]
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if result := database.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		log.Warn().Err(result.Error).Interface("user_id", userID).Msg("User not found when updating profile")
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"error":   "Not found",
			"message": "User not found",
		})
		return
	}

	// Only allow updating specific fields
	var requestBody struct {
		FirstName    *string `json:"first_name"`
		LastName     *string `json:"last_name"`
		Bio          *string `json:"bio"`
		ProfileImage *string `json:"profile_image"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": err.Error(),
		})
		return
	}

	// Update fields if provided
	if requestBody.FirstName != nil {
		user.FirstName = *requestBody.FirstName
	}
	if requestBody.LastName != nil {
		user.LastName = *requestBody.LastName
	}
	if requestBody.Bio != nil {
		user.Bio = *requestBody.Bio
	}
	if requestBody.ProfileImage != nil {
		user.ProfileImage = *requestBody.ProfileImage
	}

	if result := database.DB.Save(&user); result.Error != nil {
		log.Error().Err(result.Error).Interface("user_id", userID).Msg("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to update profile",
		})
		return
	}

	log.Info().Interface("user_id", userID).Msg("User profile updated")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// Logout godoc
// @Summary Logout from the application
// @Description Invalidate the current user's tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object false "Logout options"
// @Success 200 {object} map[string]interface{} "Successfully logged out"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security BearerAuth
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	// Extract access token
	tokenString, err := extractToken(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "No token provided",
			"message": err.Error(),
		})
		return
	}

	// Handle different logout strategies
	var request struct {
		RevokeAll bool `json:"revoke_all"`
	}
	c.ShouldBindJSON(&request)

	if request.RevokeAll {
		// Revoke all refresh tokens for this user
		if err := middleware.RevokeAllUserRefreshTokens(userID.(uint)); err != nil {
			log.Error().Err(err).Interface("user_id", userID).Msg("Failed to revoke all tokens")
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"error":   "Server error",
				"message": "Failed to revoke all tokens",
			})
			return
		}
		log.Info().Interface("user_id", userID).Msg("All refresh tokens revoked")
	}

	// Blacklist the current access token
	// Parse token to get expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(middleware.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		log.Warn().Err(err).Msg("Invalid token during logout")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Invalid token",
			"message": "The provided token is invalid or malformed",
		})
		return
	}

	// Get expiration time from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Error().Msg("Failed to parse token claims during logout")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Token parsing failed",
			"message": "An error occurred while processing the token",
		})
		return
	}

	var expiresAt time.Time
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	} else {
		expiresAt = time.Now().Add(time.Hour * 24) // Default to 24 hours if unable to extract
	}

	// Use a transaction for checking and creating the blacklisted token
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Check if token is already blacklisted
		var count int64
		if err := tx.Model(&models.BlacklistedToken{}).Where("token = ?", tokenString).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check token status: %w", err)
		}

		if count > 0 {
			// Token already blacklisted, no error needed
			return nil
		}

		// Add token to blacklist
		blacklistedToken := models.BlacklistedToken{
			Token:     tokenString,
			ExpiresAt: expiresAt,
		}

		if err := tx.Create(&blacklistedToken).Error; err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}

		return nil
	})

	// Check for transaction errors
	if err != nil {
		log.Error().Err(err).Msg("Database error during logout")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database operation failed",
			"message": "An error occurred while processing your logout request",
		})
		return
	}

	log.Info().Interface("user_id", userID).Msg("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully logged out",
	})
}

// Helper function to extract token from request
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}
