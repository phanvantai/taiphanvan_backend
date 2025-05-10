package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/rs/zerolog/log"
)

// Maximum file size (5MB)
const maxFileSize = 5 * 1024 * 1024

// Allowed file types
var allowedUploadFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".svg":  true,
	".pdf":  true,
}

// UploadFile godoc
// @Summary Upload a file for editor use
// @Description Upload a file that can be used in the editor when creating or editing posts
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload (JPG, JPEG, PNG, WEBP, GIF, SVG, PDF, max 5MB)"
// @Success 200 {object} models.SwaggerFileUploadResponse "File uploaded successfully"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /files/upload [post]
func UploadFile(c *gin.Context) {
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
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "No file uploaded or invalid file",
		})
		return
	}

	// Check file size
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "File too large",
			"message": "File must be less than 5MB",
		})
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadFileTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid file type",
			"message": "Only JPG, JPEG, PNG, WEBP, GIF, SVG, and PDF files are allowed",
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

	// Upload the file to Cloudinary
	fileURL, err := cloudinaryService.UploadEditorFile(c.Request.Context(), file, userID.(uint))
	if err != nil {
		log.Error().Err(err).Interface("user_id", userID).Msg("Failed to upload file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Upload failed",
			"message": "Failed to upload file",
		})
		return
	}

	log.Info().Interface("user_id", userID).Str("file_url", fileURL).Msg("File uploaded successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File uploaded successfully",
		"data": gin.H{
			"file_url": fileURL,
		},
	})
}

// DeleteFile godoc
// @Summary Delete a file uploaded for editor use
// @Description Delete a file that was previously uploaded for use in the editor
// @Tags Files
// @Accept json
// @Produce json
// @Param request body models.SwaggerDeleteFileRequest true "File URL to delete"
// @Success 200 {object} models.SwaggerStandardResponse "File deleted successfully"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /files/delete [post]
func DeleteFile(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	// Parse request body
	var request models.DeleteFileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "Invalid request format",
		})
		return
	}

	// Validate file URL
	if request.FileURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid input",
			"message": "File URL is required",
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

	// Delete the file from Cloudinary
	if err := cloudinaryService.DeleteImage(c.Request.Context(), request.FileURL); err != nil {
		log.Error().Err(err).Str("file_url", request.FileURL).Msg("Failed to delete file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Delete failed",
			"message": "Failed to delete file",
		})
		return
	}

	log.Info().Str("file_url", request.FileURL).Msg("File deleted successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File deleted successfully",
	})
}
