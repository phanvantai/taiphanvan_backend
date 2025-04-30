package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/phanvantai/personal_blog_backend/internal/database"
	"github.com/phanvantai/personal_blog_backend/internal/middleware"
	"github.com/phanvantai/personal_blog_backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register creates a new user account
func Register(c *gin.Context) {
	var request models.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists - use Count instead of First to avoid "record not found" error
	var count int64
	database.DB.Model(&models.User{}).Where("email = ?", request.Email).Or("username = ?", request.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login authenticates a user and returns a JWT token
func Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user by email
	var user models.User
	if result := database.DB.Where("email = ?", request.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare passwords
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, expiresIn, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	})
}

// GetProfile returns the current user's profile
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if result := database.DB.Select("id, username, email, first_name, last_name, bio, role, profile_image, created_at, updated_at").Where("id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the user's profile information
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if result := database.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	database.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// Logout invalidates the current JWT token
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No token provided",
			"message": "Authorization header is required",
		})
		return
	}

	// Extract token from header
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid authorization format",
			"message": "Authorization header format must be Bearer {token}",
		})
		return
	}

	tokenString := parts[1]

	// Parse token to get expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid token",
			"message": "The provided token is invalid or malformed",
		})
		return
	}

	// Get expiration time from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse token claims",
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
			// No error, but will handle this case after transaction
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database operation failed",
			"message": "An error occurred while processing your logout request",
		})
		return
	}

	// Check if token was already blacklisted (outside transaction to avoid DB errors)
	var count int64
	database.DB.Model(&models.BlacklistedToken{}).Where("token = ?", tokenString).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully logged out",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
