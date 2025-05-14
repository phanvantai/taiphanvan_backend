package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/rs/zerolog/log"
)

// Maximum file size (5MB)
const maxCoverSize = 5 * 1024 * 1024

// Allowed file types
var allowedCoverFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// UploadPostCover godoc
// UploadPostCover handles the request
func UploadPostCover(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized", "Authentication required"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Only allow the author to update the post cover
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden", "Only the author can update the post cover"))
		return
	}

	// Get the file from the request
	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "No file uploaded or invalid file"))
		return
	}

	// Check file size
	if file.Size > maxCoverSize {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("File too large", "Cover image must be less than 5MB"))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedCoverFileTypes[ext] {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid file type", "Only JPG, JPEG, PNG, and WEBP files are allowed"))
		return
	}

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService(middleware.AppConfig.Cloudinary)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cloudinary service")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to initialize upload service"))
		return
	}

	// Store the old cover URL for later deletion
	oldCover := post.Cover

	// Upload the file to Cloudinary
	imageURL, err := cloudinaryService.UploadPostCover(c.Request.Context(), file, uint(postID))
	if err != nil {
		log.Error().Err(err).Uint64("post_id", postID).Msg("Failed to upload cover")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Upload failed", "Failed to upload cover image"))
		return
	}

	// Update post's cover in the database
	post.Cover = imageURL
	if result := database.DB.Save(&post); result.Error != nil {
		log.Error().Err(result.Error).Uint64("post_id", postID).Msg("Failed to update post cover")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to update post cover"))
		return
	}

	// Now that the database update is successful, delete the old cover if it exists
	if oldCover != "" {
		if err := cloudinaryService.DeleteImage(c.Request.Context(), oldCover); err != nil {
			log.Warn().Err(err).Str("cover_url", oldCover).Msg("Failed to delete old cover image")
			// Continue even if deletion fails as this is not critical for the user experience
		} else {
			log.Info().Str("cover_url", oldCover).Msg("Old cover image deleted successfully")
		}
	}

	log.Info().Uint64("post_id", postID).Str("image_url", imageURL).Msg("Post cover updated")
	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"cover": imageURL,
	}, "Cover uploaded successfully"))
}

// DeletePostCover godoc
// DeletePostCover handles the request
func DeletePostCover(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized", "Authentication required"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Only allow the author to delete the post cover
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden", "Only the author can delete the post cover"))
		return
	}

	// If post has no cover, return success
	if post.Cover == "" {
		c.JSON(http.StatusOK, models.NewSuccessResponse(struct{}{}, "Post has no cover to delete"))
		return
	}

	// Store the cover URL for later deletion
	coverToDelete := post.Cover

	// Update post in the database first
	post.Cover = ""
	if result := database.DB.Save(&post); result.Error != nil {
		log.Error().Err(result.Error).Uint64("post_id", postID).Msg("Failed to update post")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to update post"))
		return
	}

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService(middleware.AppConfig.Cloudinary)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cloudinary service")
		// We've already updated the database, so we'll just log this error and continue
		log.Warn().Err(err).Str("cover_url", coverToDelete).Msg("Failed to initialize Cloudinary service for deletion")
	} else {
		// Delete the cover from Cloudinary
		if err := cloudinaryService.DeleteImage(c.Request.Context(), coverToDelete); err != nil {
			log.Warn().Err(err).Str("cover_url", coverToDelete).Msg("Failed to delete cover image from storage")
			// Continue since the database update was successful
		} else {
			log.Info().Str("cover_url", coverToDelete).Msg("Cover image deleted from storage successfully")
		}
	}

	log.Info().Uint64("post_id", postID).Msg("Post cover deleted")
	c.JSON(http.StatusOK, models.NewSuccessResponse(struct{}{}, "Cover deleted successfully"))
}
