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
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized", "Authentication required"))
		return
	}

	// Get the file from the request
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "No file uploaded or invalid file"))
		return
	}

	// Check file size
	if file.Size > maxAvatarSize {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("File too large", "Avatar image must be less than 2MB"))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedFileTypes[ext] {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid file type", "Only JPG, JPEG, and PNG files are allowed"))
		return
	}

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService(middleware.AppConfig.Cloudinary)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cloudinary service")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to initialize upload service"))
		return
	}

	// Get current user data to check if they already have an avatar
	var user models.User
	if result := database.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		log.Error().Err(result.Error).Interface("user_id", userID).Msg("Failed to find user")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to retrieve user profile"))
		return
	}

	// Store the old profile image URL for later deletion
	oldProfileImage := user.ProfileImage

	// Upload the file to Cloudinary
	imageURL, err := cloudinaryService.UploadAvatar(c.Request.Context(), file, userID.(uint))
	if err != nil {
		log.Error().Err(err).Interface("user_id", userID).Msg("Failed to upload avatar")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Upload failed", "Failed to upload avatar image"))
		return
	}

	// Update user's profile image in the database
	user.ProfileImage = imageURL
	if result := database.DB.Save(&user); result.Error != nil {
		log.Error().Err(result.Error).Interface("user_id", userID).Msg("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to update profile image"))
		return
	}

	// Now that the database update is successful, delete the old avatar if it exists
	if oldProfileImage != "" {
		if err := cloudinaryService.DeleteImage(c.Request.Context(), oldProfileImage); err != nil {
			log.Warn().Err(err).Str("profile_image_url", oldProfileImage).Msg("Failed to delete old avatar image")
			// Continue even if deletion fails as this is not critical for the user experience
		} else {
			log.Info().Str("profile_image_url", oldProfileImage).Msg("Old avatar image deleted successfully")
		}
	}

	log.Info().Interface("user_id", userID).Str("image_url", imageURL).Msg("User avatar updated")
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"profile_image": imageURL,
	}, "Avatar uploaded successfully"))
}
