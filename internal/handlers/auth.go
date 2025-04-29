package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/personal_blog_backend/internal/database"
	"github.com/phanvantai/personal_blog_backend/internal/middleware"
	"github.com/phanvantai/personal_blog_backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account
func Register(c *gin.Context) {
	var request models.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if result := database.DB.Where("email = ?", request.Email).Or("username = ?", request.Username).First(&existingUser); result.RowsAffected > 0 {
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
