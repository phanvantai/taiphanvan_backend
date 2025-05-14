package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// buildPostQuery is a generic helper function to build a query for posts with common preloads
func buildPostQuery[T any](query *gorm.DB) *gorm.DB {
	return query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Preload("Tags").Order("created_at DESC")
}

// getPaginatedPosts is a generic helper function to fetch paginated posts
func getPaginatedPosts[T any](page, limit int, queryBuilder func(*gorm.DB) *gorm.DB) ([]T, int64, error) {
	var items []T
	var total int64
	offset := (page - 1) * limit

	query := database.DB.Model(new(T))
	query = queryBuilder(query)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// createPaginationMeta creates a pagination metadata object
func createPaginationMeta(page, limit int, total int64) gin.H {
	return gin.H{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"lastPage": (int(total) + limit - 1) / limit,
	}
}

// GetPosts godoc
// GetPosts handles the request
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tag := c.Query("tag")
	status := c.Query("status")

	posts, total, err := getPaginatedPosts[models.Post](page, limit, func(query *gorm.DB) *gorm.DB {
		query = buildPostQuery[models.Post](query)

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

		return query
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to fetch posts"))
		return
	}

	response := gin.H{
		"posts": posts,
		"meta":  createPaginationMeta(page, limit, total),
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Posts retrieved successfully"))
}

// GetPostBySlug godoc
// GetPostBySlug handles the request
func GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var post models.Post
	if err := database.DB.Where("slug = ?", slug).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Preload("Tags").First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Increment view count in a separate goroutine to not block the response
	go func(postID uint) {
		if err := database.DB.Model(&models.Post{}).Where("id = ?", postID).
			Update("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
			log.Error().Err(err).Uint("post_id", postID).Msg("Failed to increment view count")
		}
	}(post.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(post, "Post retrieved successfully"))
}

// CreatePost godoc
// CreatePost handles the request
func CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, exists := c.Get("userRole")

	// Improved permission check with better logging and error handling
	userRole, ok := role.(string)
	if !exists || !ok || (userRole != "admin" && userRole != "editor") {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Permission denied",
			"Only administrators and editors can create posts",
		))
		return
	}

	var requestBody models.CreatePostRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
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
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid input",
				"Scheduled publish date must be in the future",
			))
			return
		}
	}

	tx := database.DB.Begin()

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to create post"))
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
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to process tags"))
				return
			}

			// Associate tag with post
			if err := tx.Model(&post).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to associate tags"))
				return
			}
		}
	}

	tx.Commit()

	// Reload post with tags
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name")
	}).First(&post, post.ID)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(post, "Post created successfully"))
}

// UpdatePost godoc
// UpdatePost handles the request
func UpdatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Only the author can update the post
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Permission denied",
			"Only the author can update this post",
		))
		return
	}

	var requestBody models.UpdatePostRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
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
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid input",
				"Invalid status. Allowed values: draft, published, archived, scheduled",
			))
			return
		}

		// Update the status
		post.Status = *requestBody.Status

		// Handle scheduled posts
		if *requestBody.Status == models.PostStatusScheduled && requestBody.PublishAt != nil {
			// Validate the publish date is in the future
			if requestBody.PublishAt.Before(time.Now()) {
				c.JSON(http.StatusBadRequest, models.NewErrorResponse(
					"Invalid input",
					"Scheduled publish date must be in the future",
				))
				return
			}
		}
	}

	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to update post"))
		return
	}

	// Update tags if provided
	if len(requestBody.Tags) > 0 {
		// Clear existing tags
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to update tags"))
			return
		}

		for _, tagName := range requestBody.Tags {
			var tag models.Tag
			tagName = strings.TrimSpace(tagName)

			// Find or create tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to process tags"))
				return
			}

			// Associate tag with post
			if err := tx.Model(&post).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to associate tags"))
				return
			}
		}
	}

	tx.Commit()

	// Reload post with tags
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(post, "Post updated successfully"))
}

// DeletePost godoc
// DeletePost handles the request
func DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Check if user is the author or an admin
	role, exists := c.Get("userRole")
	userRole, ok := role.(string)
	isAdmin := exists && ok && userRole == "admin"

	if post.UserID != userID.(uint) && !isAdmin {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Permission denied",
			"You don't have permission to delete this post",
		))
		return
	}

	// Delete post (soft delete because of gorm.DeletedAt field)
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to delete post"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse[any](nil, "Post deleted successfully"))
}

// PublishPost godoc
// PublishPost handles the request
func PublishPost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Only the author can publish the post
	if post.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Permission denied",
			"Only the author can publish this post",
		))
		return
	}

	// Check if post is already published
	if post.Status == models.PostStatusPublished {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid operation", "Post is already published"))
		return
	}

	// Set status to published
	post.Status = models.PostStatusPublished

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to publish post"))
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(post, "Post published successfully"))
}

// UnpublishPost godoc
// UnpublishPost handles the request
func UnpublishPost(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	// Check if user is the author or an admin
	role, exists := c.Get("userRole")
	userRole, ok := role.(string)
	isAdmin := exists && ok && userRole == "admin"

	if post.UserID != userID.(uint) && !isAdmin {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Permission denied",
			"You don't have permission to unpublish this post",
		))
		return
	}

	// Check if post is already a draft
	if post.Status == models.PostStatusDraft {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid operation", "Post is already unpublished"))
		return
	}

	// Set status to draft
	post.Status = models.PostStatusDraft

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to unpublish post"))
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(post, "Post unpublished successfully"))
}

// SetPostStatus godoc
// SetPostStatus handles the request
func SetPostStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	roleInterface, exists := c.Get("userRole")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid post ID"))
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Post not found"))
		return
	}

	var requestBody models.SetPostStatusRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
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
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				"Permission denied",
				"Only the author can publish this post",
			))
			return
		}
	} else if requestBody.Status == models.PostStatusDraft {
		// Only admin or author can unpublish a post
		if !isAuthor && !isAdmin {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				"Permission denied",
				"Only the author or administrators can unpublish this post",
			))
			return
		}
	} else {
		// For other status changes (archived, scheduled), only the author can do this
		if !isAuthor {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				"Permission denied",
				"Only the author can change the status of this post",
			))
			return
		}
	}

	// Validate status
	switch requestBody.Status {
	case models.PostStatusDraft, models.PostStatusPublished, models.PostStatusArchived, models.PostStatusScheduled:
		// Valid status
	default:
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid input",
			"Invalid status. Allowed values: draft, published, archived, scheduled",
		))
		return
	}

	// For scheduled posts, validate the publish date
	if requestBody.Status == models.PostStatusScheduled {
		// For scheduled posts, check if we have a publish date
		if requestBody.PublishAt == nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid input",
				"PublishAt date is required for scheduled posts",
			))
			return
		}

		// Validate the publish date is in the future
		if requestBody.PublishAt.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid input",
				"PublishAt date must be in the future",
			))
			return
		}
	}

	// Update the status
	post.Status = requestBody.Status

	if err := database.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to update post status"))
		return
	}

	// Reload post with tags and user
	database.DB.Preload("Tags").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).First(&post, post.ID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(post, "Post status updated successfully"))
}

// GetMyPosts godoc
// GetMyPosts handles the request
func GetMyPosts(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"Unauthorized",
			"Authentication required",
		))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, total, err := getPaginatedPosts[models.Post](page, limit, func(query *gorm.DB) *gorm.DB {
		return buildPostQuery[models.Post](query).Where("user_id = ?", userID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Server error",
			"Failed to fetch posts",
		))
		return
	}

	response := gin.H{
		"posts": posts,
		"meta":  createPaginationMeta(page, limit, total),
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Your posts retrieved successfully"))
}

// IncrementPostViewCount godoc
// IncrementPostViewCount handles the request
func IncrementPostViewCount(c *gin.Context) {
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

	// Increment view count
	if err := database.DB.Model(&post).Update("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to update view count"))
		return
	}

	data := gin.H{
		"post_id":    post.ID,
		"view_count": post.ViewCount + 1, // Return the incremented count
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(data, "View count incremented successfully"))
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
