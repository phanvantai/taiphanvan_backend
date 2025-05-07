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
// @Summary Upload post cover image
// @Description Upload a new cover image for a post
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Post ID"
// @Param cover formData file true "Cover image file (JPG, JPEG, PNG, WEBP, max 5MB)"
// @Success 200 {object} map[string]interface{} "Cover uploaded successfully"
// @Success 200 {object} map[string]interface{} "Example response" {{"status":"success","message":"Cover uploaded successfully","data":{"cover":"https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg"}}}
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security BearerAuth
// @Router /posts/{id}/cover [post]
func UploadPostCover(c *gin.Context) {
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

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "Invalid post ID",
		})
		return
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"error":   "Not found",
			"message": "Post not found",
		})
		return
	}

	// Check if user is the author or an admin
	role, _ := c.Get("userRole")
	if post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Forbidden",
			"message": "You don't have permission to update this post",
		})
		return
	}

	// Get the file from the request
	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "No file uploaded or invalid file",
		})
		return
	}

	// Check file size
	if file.Size > maxCoverSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "File too large",
			"message": "Cover image must be less than 5MB",
		})
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedCoverFileTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid file type",
			"message": "Only JPG, JPEG, PNG, and WEBP files are allowed",
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

	// If post already has a cover, delete the old one
	if post.Cover != "" {
		if err := cloudinaryService.DeleteImage(c.Request.Context(), post.Cover); err != nil {
			log.Warn().Err(err).Str("cover_url", post.Cover).Msg("Failed to delete old cover image")
			// Continue with the upload even if deletion fails
		}
	}

	// Upload the file to Cloudinary
	imageURL, err := cloudinaryService.UploadPostCover(c.Request.Context(), file, uint(postID))
	if err != nil {
		log.Error().Err(err).Uint64("post_id", postID).Msg("Failed to upload cover")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Upload failed",
			"message": "Failed to upload cover image",
		})
		return
	}

	// Update post's cover in the database
	post.Cover = imageURL
	if result := database.DB.Save(&post); result.Error != nil {
		log.Error().Err(result.Error).Uint64("post_id", postID).Msg("Failed to update post cover")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to update post cover",
		})
		return
	}

	log.Info().Uint64("post_id", postID).Str("image_url", imageURL).Msg("Post cover updated")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cover uploaded successfully",
		"data": gin.H{
			"cover": imageURL,
		},
	})
}

// DeletePostCover godoc
// @Summary Delete post cover image
// @Description Remove the cover image from a post
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{} "Cover deleted successfully"
// @Success 200 {object} map[string]interface{} "Example response" {{"status":"success","message":"Cover deleted successfully"}}
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security BearerAuth
// @Router /posts/{id}/cover [delete]
func DeletePostCover(c *gin.Context) {
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

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "Invalid post ID",
		})
		return
	}

	// Find the post
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"error":   "Not found",
			"message": "Post not found",
		})
		return
	}

	// Check if user is the author or an admin
	role, _ := c.Get("userRole")
	if post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Forbidden",
			"message": "You don't have permission to update this post",
		})
		return
	}

	// If post has no cover, return success
	if post.Cover == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Post has no cover to delete",
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
			"message": "Failed to initialize service",
		})
		return
	}

	// Delete the cover from Cloudinary
	if err := cloudinaryService.DeleteImage(c.Request.Context(), post.Cover); err != nil {
		log.Error().Err(err).Str("cover_url", post.Cover).Msg("Failed to delete cover image")
		// Continue with the database update even if Cloudinary deletion fails
	}

	// Update post in the database
	post.Cover = ""
	if result := database.DB.Save(&post); result.Error != nil {
		log.Error().Err(result.Error).Uint64("post_id", postID).Msg("Failed to update post")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to update post",
		})
		return
	}

	log.Info().Uint64("post_id", postID).Msg("Post cover deleted")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cover deleted successfully",
	})
}
