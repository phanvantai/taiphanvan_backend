package handlers

import (
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
// @Description Returns a paginated list of blog posts with optional tag filtering
// @Tags Posts
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10)"
// @Param tag query string false "Filter posts by tag name"
// @Success 200 {object} map[string]interface{} "List of posts with pagination metadata"
// @Success 200 {object} map[string]interface{} "Example response" {{"posts":[{"id":1,"title":"Sample Post","slug":"sample-post","content":"This is a sample post content","excerpt":"Sample excerpt","cover":"https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg","user_id":1,"user":{"id":1,"username":"johndoe","first_name":"John","last_name":"Doe","profile_image":"https://example.com/avatar.jpg"},"tags":[{"id":1,"name":"technology"}],"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-02T12:00:00Z","published_at":"2023-01-03T12:00:00Z"}],"meta":{"page":1,"limit":10,"total":1,"lastPage":1}}}
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /posts [get]
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tag := c.Query("tag")

	offset := (page - 1) * limit
	var posts []models.Post
	query := database.DB.Model(&models.Post{}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, first_name, last_name, profile_image")
	}).Preload("Tags").Order("created_at DESC")

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
// @Failure 404 {object} map[string]interface{} "Post not found"
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
// @Param post body object true "Post details" {{"title":"My New Post","content":"This is the content of my new post","excerpt":"A short excerpt","cover":"","tags":["technology","programming"],"published":true}}
// @Success 201 {object} models.Post "Created post"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security BearerAuth
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	var requestBody struct {
		Title     string   `json:"title" binding:"required"`
		Content   string   `json:"content" binding:"required"`
		Excerpt   string   `json:"excerpt"`
		Cover     string   `json:"cover"`
		Tags      []string `json:"tags"`
		Published bool     `json:"published"`
	}

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

	if requestBody.Published {
		now := time.Now()
		post.PublishedAt = &now
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
// @Param post body object true "Post details" {{"title":"Updated Post Title","content":"Updated content","excerpt":"Updated excerpt","cover":"https://example.com/updated-cover.jpg","tags":["technology","programming","updated"],"published":true}}
// @Success 200 {object} models.Post "Updated post"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Failure 500 {object} map[string]interface{} "Server error"
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

	// Check if user is the author or an admin
	role, _ := c.Get("userRole")
	if post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this post"})
		return
	}

	var requestBody struct {
		Title     *string  `json:"title"`
		Content   *string  `json:"content"`
		Excerpt   *string  `json:"excerpt"`
		Cover     *string  `json:"cover"`
		Tags      []string `json:"tags"`
		Published *bool    `json:"published"`
	}

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
	if requestBody.Published != nil {
		if *requestBody.Published && post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		} else if !*requestBody.Published {
			post.PublishedAt = nil
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
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Failure 500 {object} map[string]interface{} "Server error"
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
	role, _ := c.Get("userRole")
	if post.UserID != userID.(uint) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this post"})
		return
	}

	// Delete post (soft delete because of gorm.DeletedAt field)
	if err := database.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
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
