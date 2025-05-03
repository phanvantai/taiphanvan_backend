package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"gorm.io/gorm"
)

// GetCommentsByPostID returns all comments for a specific post
func GetCommentsByPostID(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("postID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var comments []models.Comment
	if err := database.DB.Where("post_id = ?", postID).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Order("created_at DESC").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// CreateComment adds a new comment to a post
func CreateComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := strconv.ParseUint(c.Param("postID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Check if post exists
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var requestBody struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.Comment{
		Content: requestBody.Content,
		PostID:  uint(postID),
		UserID:  userID.(uint),
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Reload comment with user info
	database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&comment, comment.ID)

	c.JSON(http.StatusCreated, comment)
}

// UpdateComment updates an existing comment
func UpdateComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Check if user is the author of the comment or an admin
	role, _ := c.Get("userRole")
	if comment.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this comment"})
		return
	}

	var requestBody struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.Content = requestBody.Content

	if err := database.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	// Reload comment with user info
	database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&comment, comment.ID)

	c.JSON(http.StatusOK, comment)
}

// DeleteComment removes a comment
func DeleteComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Check if user is the author of the comment, post author, or an admin
	role, _ := c.Get("userRole")
	var post models.Post
	database.DB.First(&post, comment.PostID)

	if comment.UserID != userID.(uint) && post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this comment"})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
