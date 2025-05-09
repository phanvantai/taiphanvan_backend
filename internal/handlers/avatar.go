package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/rs/zerolog/log"
)

// Maximum file size (2MB)
const maxAvatarSize = 2 * 1024 * 1024

// Allowed file types
var allowedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload a new avatar image for the current user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file (JPG, JPEG, PNG, max 2MB)"
// @Success 200 {object} models.SwaggerAvatarResponse "Avatar uploaded successfully"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /profile/avatar [post]
func UploadAvatar(c *gin.Context) {
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

	// Get the file from the request
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "No file uploaded or invalid file",
		})
		return
	}

	// Check file size
	if file.Size > maxAvatarSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "File too large",
			"message": "Avatar image must be less than 2MB",
		})
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedFileTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid file type",
			"message": "Only JPG, JPEG, and PNG files are allowed",
		})
		return
	}

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService(middleware.AppConfig.Cloudinary)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cloudinary service")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Server error",
			"message": "Failed to initialize upload service",
		})
		return
	}

	// Get current user data to check if they already have an avatar
	var user models.User
	if result := database.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		log.Error().Err(result.Error).Interface("user_id", userID).Msg("Failed to find user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to retrieve user profile",
		})
		return
	}

	// If user already has a profile image, delete the old one
	if user.ProfileImage != "" {
		if err := cloudinaryService.DeleteImage(c.Request.Context(), user.ProfileImage); err != nil {
			log.Warn().Err(err).Str("profile_image_url", user.ProfileImage).Msg("Failed to delete old avatar image")
			// Continue with the upload even if deletion fails
		}
	}

	// Upload the file to Cloudinary
	imageURL, err := cloudinaryService.UploadAvatar(c.Request.Context(), file, userID.(uint))
	if err != nil {
		log.Error().Err(err).Interface("user_id", userID).Msg("Failed to upload avatar")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Upload failed",
			"message": "Failed to upload avatar image",
		})
		return
	}

	// Update user's profile image in the database
	user.ProfileImage = imageURL
	if result := database.DB.Save(&user); result.Error != nil {
		log.Error().Err(result.Error).Interface("user_id", userID).Msg("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to update profile image",
		})
		return
	}

	log.Info().Interface("user_id", userID).Str("image_url", imageURL).Msg("User avatar updated")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Avatar uploaded successfully",
		"data": gin.H{
			"profile_image": imageURL,
		},
	})
}
