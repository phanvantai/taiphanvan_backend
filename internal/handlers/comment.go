package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"gorm.io/gorm"
)

// GetCommentsByPostID godoc
// @Summary Get comments for a post
// @Description Returns all comments for a specific post
// @Tags Comments
// @Produce json
// @Param postID path int true "Post ID"
// @Success 200 {object} models.SwaggerStandardResponse "List of comments"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /posts/{postID}/comments [get]
func GetCommentsByPostID(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var comments []models.Comment
	if err := database.DB.Where("post_id = ?", postID).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Order("created_at DESC").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to fetch comments"))
		return
	}

	// Check if user is authenticated to include vote status
	var userID uint
	var isAuthenticated bool
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
		isAuthenticated = true
	}

	// If the user is authenticated, include their vote status for each comment
	type CommentWithVote struct {
		*models.Comment
		UserVote int8 `json:"user_vote"`
	}

	var result []CommentWithVote
	for _, comment := range comments {
		commentWithVote := CommentWithVote{
			Comment:  &comment,
			UserVote: 0,
		}

		if isAuthenticated {
			voteType, err := GetUserVote(comment.ID, userID)
			if err == nil {
				commentWithVote.UserVote = int8(voteType)
			}
		}

		result = append(result, commentWithVote)
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result, "Comments retrieved successfully"))
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Adds a new comment to a post
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param comment body models.CreateCommentRequest true "Comment content"
// @Success 201 {object} models.SwaggerStandardResponse "Created comment"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id}/comments [post]
func CreateComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	// Check if post exists
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	var requestBody models.CreateCommentRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
		return
	}

	comment := models.Comment{
		Content: requestBody.Content,
		PostID:  uint(postID),
		UserID:  userID.(uint),
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to create comment"))
		return
	}

	// Reload comment with user info
	database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&comment, comment.ID)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(comment, "Comment created successfully"))
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Updates an existing comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentID path int true "Comment ID"
// @Param comment body models.UpdateCommentRequest true "Updated comment content"
// @Success 200 {object} models.SwaggerStandardResponse "Updated comment"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Comment not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /comments/{commentID} [put]
func UpdateComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid comment ID"))
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Comment not found"))
		return
	}

	// Check if user is the author of the comment or an admin
	role, _ := c.Get("userRole")
	if comment.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden", "You don't have permission to edit this comment"))
		return
	}

	var requestBody models.UpdateCommentRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
		return
	}

	comment.Content = requestBody.Content

	if err := database.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to update comment"))
		return
	}

	// Reload comment with user info
	database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&comment, comment.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(comment, "Comment updated successfully"))
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Removes a comment from a post
// @Tags Comments
// @Produce json
// @Param commentID path int true "Comment ID"
// @Success 200 {object} models.SwaggerStandardResponse "Success message"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Comment not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /comments/{commentID} [delete]
func DeleteComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid comment ID"))
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Comment not found"))
		return
	}

	// Check if user is the author of the comment, post author, or an admin
	role, _ := c.Get("userRole")
	var post models.Post
	database.DB.First(&post, comment.PostID)

	if comment.UserID != userID.(uint) && post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, models.NewErrorResponse("Forbidden", "You don't have permission to delete this comment"))
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Database error", "Failed to delete comment"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(struct{}{}, "Comment deleted successfully"))
}
