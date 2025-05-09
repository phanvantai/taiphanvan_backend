package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"gorm.io/gorm"
)

// GetPosts godoc
// @Summary Get list of blog posts
// @Description Returns a paginated list of blog posts with optional tag and status filtering
// @Tags Posts
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10)"
// @Param tag query string false "Filter posts by tag name"
// @Param status query string false "Filter posts by status (draft, published, archived, scheduled)"
// @Success 200 {object} models.SwaggerPostsResponse "List of posts with pagination metadata"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /posts [get]
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tag := c.Query("tag")
	status := c.Query("status")

	offset := (page - 1) * limit
	var posts []models.Post
	query := database.DB.Model(&models.Post{}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Preload("Tags").Order("created_at DESC")

	// Default to showing only published posts for public API
	if status == "" {
		query = query.Where("status = ?", models.PostStatusPublished)
	} else {
		// If specific status is requested, filter by that status
		query = query.Where("status = ?", status)
	}

	// Filter by tag if specified
	if tag != "" {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.name = ?", tag)
	}

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"meta": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"lastPage": (int(total) + limit - 1) / limit,
		},
	})
}

// GetPostBySlug godoc
// @Summary Get a blog post by slug
// @Description Returns a single blog post by its slug
// @Tags Posts
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} models.Post "Post details"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Router /posts/slug/{slug} [get]
func GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var post models.Post
	if err := database.DB.Where("slug = ?", slug).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Preload("Tags").First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// CreatePost godoc
// @Summary Create a new blog post
// @Description Creates a new blog post with the provided details
// @Tags Posts
// @Accept json
// @Produce json
// @Param request body models.CreatePostRequest true "Post details"
// @Success 201 {object} models.Post "Created post"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, exists := c.Get("userRole")

	// Improved permission check with better logging and error handling
	userRole, ok := role.(string)
	if !exists || !ok || (userRole != "admin" && userRole != "editor") {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Permission denied",
			"message": "Only administrators and editors can create posts",
			"debug":   fmt.Sprintf("User role: %v, exists: %v, ok: %v", role, exists, ok),
		})
		return
	}

	var requestBody models.CreatePostRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a slug from the title
	slug := generateSlug(requestBody.Title)

	// Check if slug already exists
	var existingPost models.Post
	if result := database.DB.Where("slug = ?", slug).First(&existingPost); result.RowsAffected > 0 {
		// Append a random suffix to make the slug unique
		slug = slug + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Create the post
	post := models.Post{
		Title:   requestBody.Title,
		Content: requestBody.Content,
		Excerpt: requestBody.Excerpt,
		Cover:   requestBody.Cover,
		Slug:    slug,
		UserID:  userID.(uint),
	}

	// Set status (default to draft if not specified)
	if requestBody.Status != "" {
		post.Status = requestBody.Status
	} else {
		post.Status = models.PostStatusDraft
	}

	// Handle scheduled posts
	if post.Status == models.PostStatusScheduled && requestBody.PublishAt != nil {
		// Validate the publish date is in the future
		if requestBody.PublishAt.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled publish date must be in the future"})
			return
		}
	}

	tx := database.DB.Begin()

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// Add tags
	if len(requestBody.Tags) > 0 {
		for _, tagName := range requestBody.Tags {
			var tag models.Tag
			tagName = strings.TrimSpace(tagName)

			// Find or create tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tags"})
				return
			}

			// Associate tag with post
			if err := tx.Model(&post).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
				return
			}
		}
	}

	tx.Commit()

	// Reload post with tags
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name")
	}).First(&post, post.ID)

	c.JSON(http.StatusCreated, post)
}

// UpdatePost godoc
// @Summary Update an existing blog post
// @Description Updates a blog post with the provided details
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param post body models.UpdatePostRequest true "Post details"
// @Success 200 {object} models.Post "Updated post"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Only the author can update the post
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the author can update this post"})
		return
	}

	var requestBody models.UpdatePostRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	// Update fields if provided
	if requestBody.Title != nil {
		post.Title = *requestBody.Title
		// Update slug only if title changes
		post.Slug = generateSlug(*requestBody.Title)
	}
	if requestBody.Content != nil {
		post.Content = *requestBody.Content
	}
	if requestBody.Excerpt != nil {
		post.Excerpt = *requestBody.Excerpt
	}
	if requestBody.Cover != nil {
		post.Cover = *requestBody.Cover
	}

	// Handle status update
	if requestBody.Status != nil {
		// Validate status
		switch *requestBody.Status {
		case models.PostStatusDraft, models.PostStatusPublished, models.PostStatusArchived, models.PostStatusScheduled:
			// Valid status
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Allowed values: draft, published, archived, scheduled"})
			return
		}

		// Update the status
		post.Status = *requestBody.Status

		// Handle scheduled posts
		if *requestBody.Status == models.PostStatusScheduled && requestBody.PublishAt != nil {
			// Validate the publish date is in the future
			if requestBody.PublishAt.Before(time.Now()) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled publish date must be in the future"})
				return
			}
		}
	}

	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	// Update tags if provided
	if len(requestBody.Tags) > 0 {
		// Clear existing tags
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags"})
			return
		}

		for _, tagName := range requestBody.Tags {
			var tag models.Tag
			tagName = strings.TrimSpace(tagName)

			// Find or create tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tags"})
				return
			}

			// Associate tag with post
			if err := tx.Model(&post).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
				return
			}
		}
	}

	tx.Commit()

	// Reload post with tags
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, post)
}

// DeletePost godoc
// @Summary Delete a blog post
// @Description Deletes a blog post by ID (soft delete)
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.SwaggerStandardResponse "Success message"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Check if user is the author or an admin
	role, exists := c.Get("userRole")
	userRole, ok := role.(string)
	isAdmin := exists && ok && userRole == "admin"

	if post.UserID != userID.(uint) && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Permission denied",
			"message": "You don't have permission to delete this post",
		})
		return
	}

	// Delete post (soft delete because of gorm.DeletedAt field)
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// PublishPost godoc
// @Summary Publish a blog post
// @Description Sets a blog post's status to published
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.Post "Published post"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id}/publish [post]
func PublishPost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Only the author can publish the post
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Permission denied",
			"message": "Only the author can publish this post",
		})
		return
	}

	// Check if post is already published
	if post.Status == models.PostStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post is already published"})
		return
	}

	// Set status to published
	post.Status = models.PostStatusPublished

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish post"})
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, post)
}

// UnpublishPost godoc
// @Summary Unpublish a blog post
// @Description Sets a blog post's status to unpublished (draft)
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.Post "Unpublished post"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id}/unpublish [post]
func UnpublishPost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Check if user is the author or an admin
	role, exists := c.Get("userRole")
	userRole, ok := role.(string)
	isAdmin := exists && ok && userRole == "admin"

	if post.UserID != userID.(uint) && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"error":   "Permission denied",
			"message": "You don't have permission to unpublish this post",
		})
		return
	}

	// Check if post is already a draft
	if post.Status == models.PostStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post is already unpublished"})
		return
	}

	// Set status to draft
	post.Status = models.PostStatusDraft

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish post"})
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, post)
}

// SetPostStatus godoc
// @Summary Set the status of a blog post
// @Description Updates a post's status to the specified value (draft, published, archived, scheduled)
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param request body models.SetPostStatusRequest true "Status details"
// @Success 200 {object} models.Post "Updated post"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 403 {object} models.SwaggerStandardResponse "Forbidden"
// @Failure 404 {object} models.SwaggerStandardResponse "Post not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/{id}/status [post]
func SetPostStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	roleInterface, exists := c.Get("userRole")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var requestBody models.SetPostStatusRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Properly handling role type assertion
	userRole, ok := roleInterface.(string)

	// Authorization check based on the requested status change and user role
	isAuthor := post.UserID == userID.(uint)
	isAdmin := exists && ok && userRole == "admin"

	// Check permissions based on the action being performed
	if requestBody.Status == models.PostStatusPublished {
		// Only the author can publish their post
		if !isAuthor {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"error":   "Permission denied",
				"message": "Only the author can publish this post",
			})
			return
		}
	} else if requestBody.Status == models.PostStatusDraft {
		// Only admin or author can unpublish a post
		if !isAuthor && !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"error":   "Permission denied",
				"message": "Only the author or administrators can unpublish this post",
			})
			return
		}
	} else {
		// For other status changes (archived, scheduled), only the author can do this
		if !isAuthor {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"error":   "Permission denied",
				"message": "Only the author can change the status of this post",
			})
			return
		}
	}

	// Validate status
	switch requestBody.Status {
	case models.PostStatusDraft, models.PostStatusPublished, models.PostStatusArchived, models.PostStatusScheduled:
		// Valid status
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Allowed values: draft, published, archived, scheduled"})
		return
	}

	// For scheduled posts, validate the publish date
	if requestBody.Status == models.PostStatusScheduled {
		// For scheduled posts, check if we have a publish date
		if requestBody.PublishAt == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PublishAt date is required for scheduled posts"})
			return
		}

		// Validate the publish date is in the future
		if requestBody.PublishAt.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PublishAt date must be in the future"})
			return
		}
	}

	// Update the status
	post.Status = requestBody.Status

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post status"})
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, post)
}

// GetMyPosts godoc
// @Summary Get the current user's blog posts
// @Description Returns a paginated list of blog posts authored by the currently authenticated user
// @Tags Posts
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10)"
// @Success 200 {object} models.SwaggerPostsResponse "List of the user's posts with pagination metadata"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /posts/me [get]
func GetMyPosts(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	offset := (page - 1) * limit
	var posts []models.Post
	query := database.DB.Model(&models.Post{}).Where("user_id = ?", userID).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, first_name, last_name, profile_image")
		}).
		Preload("Tags").
		Order("created_at DESC")

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Database error",
			"message": "Failed to fetch posts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"posts":  posts,
		"meta": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"lastPage": (int(total) + limit - 1) / limit,
		},
	})
}

// Helper function to generate slug from title
func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any character that's not alphanumeric or hyphen
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	// Replace multiple hyphens with a single hyphen
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
