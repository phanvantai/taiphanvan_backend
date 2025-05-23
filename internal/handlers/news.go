package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Default pagination values
const (
	defaultNewsPage    = 1
	defaultNewsPerPage = 10
	maxNewsPerPage     = 50
)

// GetNews godoc
// @Summary Get news articles
// @Description Returns paginated news articles with optional filtering
// @Tags News
// @Produce json
// @Param category query string false "Filter by category"
// @Param tag query string false "Filter by tag"
// @Param search query string false "Search in title and content"
// @Param page query int false "Page number, default is 1"
// @Param per_page query int false "Items per page, default is 10, max is 50"
// @Success 200 {object} models.NewsWithoutContentResponse "List of news articles with pagination (without content)"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /news [get]
func GetNews(c *gin.Context) {
	var query models.NewsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Apply default and max values for pagination
	if query.Page <= 0 {
		query.Page = defaultNewsPage
	}
	if query.PerPage <= 0 {
		query.PerPage = defaultNewsPerPage
	}
	if query.PerPage > maxNewsPerPage {
		query.PerPage = maxNewsPerPage
	}

	// Create database query
	dbQuery := database.DB.Model(&models.News{}).
		Where("status = ? AND published = ?", models.NewsStatusPublished, true)

	// Apply category filter if provided
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	// Apply tag filter if provided
	if query.Tag != "" {
		dbQuery = dbQuery.Joins("JOIN news_tags ON news_tags.news_id = news.id").
			Joins("JOIN tags ON tags.id = news_tags.tag_id").
			Where("tags.name = ?", query.Tag)
	}

	// Apply search if provided
	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("title ILIKE ? OR content ILIKE ? OR summary ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Count total items for pagination
	var totalItems int64
	if err := dbQuery.Count(&totalItems).Error; err != nil {
		log.Error().Err(err).Msg("Failed to count news articles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news articles"})
		return
	}

	// Calculate pagination values
	totalPages := (int(totalItems) + query.PerPage - 1) / query.PerPage
	offset := (query.Page - 1) * query.PerPage

	// Retrieve news with limit and offset
	var news []models.News
	if err := dbQuery.
		Order("publish_date DESC").
		Limit(query.PerPage).
		Offset(offset).
		Preload("Tags").
		Find(&news).Error; err != nil {
		log.Error().Err(err).Msg("Failed to retrieve news articles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news articles"})
		return
	}

	// Create response without content to improve performance
	var newsWithoutContent []models.NewsWithoutContent
	for _, article := range news {
		newsWithoutContent = append(newsWithoutContent, article.ToNewsWithoutContent())
	}

	response := models.NewsWithoutContentResponse{
		News:       newsWithoutContent,
		TotalItems: totalItems,
		Page:       query.Page,
		PerPage:    query.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetNewsBySlug godoc
// @Summary Get news article by slug
// @Description Returns a specific news article by its slug
// @Tags News
// @Produce json
// @Param slug path string true "News article slug"
// @Success 200 {object} models.SwaggerNewsWithContentStatus "News article with content status"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /news/slug/{slug} [get]
func GetNewsBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	var news models.News
	if err := database.DB.
		Where("slug = ? AND status = ? AND published = ?", slug, models.NewsStatusPublished, true).
		Preload("Tags").
		First(&news).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Str("slug", slug).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news article"})
		}
		return
	}

	// Check if content is truncated
	isTruncated := services.IsTruncated(news.Content)
	_, truncatedChars, _ := services.ExtractTruncationInfo(news.Content)

	// Add content status information
	contentStatus := models.ContentStatus{
		IsTruncated:    isTruncated,
		TruncatedChars: truncatedChars,
		HasFullContent: false, // We haven't fetched full content yet
	}

	c.JSON(http.StatusOK, models.NewsWithContentStatus{
		News:          news,
		ContentStatus: contentStatus,
	})
}

// GetNewsByID godoc
// @Summary Get news article by ID
// @Description Returns a specific news article by its ID
// @Tags News
// @Produce json
// @Param id path int true "News article ID"
// @Success 200 {object} models.SwaggerNewsWithContentStatus "News article with content status"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /news/{id} [get]
func GetNewsByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var news models.News
	if err := database.DB.
		Where("id = ? AND status = ? AND published = ?", id, models.NewsStatusPublished, true).
		Preload("Tags").
		First(&news).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Str("id", id).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news article"})
		}
		return
	}

	// Check if content is truncated
	isTruncated := services.IsTruncated(news.Content)
	_, truncatedChars, _ := services.ExtractTruncationInfo(news.Content)

	// Add content status information
	contentStatus := models.ContentStatus{
		IsTruncated:    isTruncated,
		TruncatedChars: truncatedChars,
		HasFullContent: false, // We haven't fetched full content yet
	}

	c.JSON(http.StatusOK, models.NewsWithContentStatus{
		News:          news,
		ContentStatus: contentStatus,
	})
}

// GetNewsCategories godoc
// @Summary Get news categories
// @Description Returns all available news categories
// @Tags News
// @Produce json
// @Success 200 {array} string "List of categories"
// @Router /news/categories [get]
func GetNewsCategories(c *gin.Context) {
	categories := []string{
		// string(models.NewsCategoryGeneral),
		// string(models.NewsCategoryBusiness),
		string(models.NewsCategoryTechnology),
		string(models.NewsCategoryScience),
		// string(models.NewsCategoryHealth),
		// string(models.NewsCategorySports),
		// string(models.NewsCategoryEntertainment),
	}

	c.JSON(http.StatusOK, categories)
}

// CreateNews godoc
// @Summary Create a news article
// @Description Create a new news article (admin only)
// @Tags News
// @Accept json
// @Produce json
// @Param news body models.CreateNewsRequest true "News article to create"
// @Success 201 {object} models.News "Created news article"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news [post]
func CreateNews(c *gin.Context) {
	// Parse request body
	var requestBody models.CreateNewsRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Generate slug from title
	newsSlug := slug.Make(requestBody.Title)

	// Check if slug already exists
	var count int64
	if err := database.DB.Model(&models.News{}).Where("slug = ?", newsSlug).Count(&count).Error; err != nil {
		log.Error().Err(err).Str("slug", newsSlug).Msg("Failed to check for existing slug")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create news article"})
		return
	}

	// If slug exists, append a timestamp
	if count > 0 {
		newsSlug = fmt.Sprintf("%s-%d", newsSlug, time.Now().Unix())
	}

	// Set publish date to now if not provided
	publishDate := time.Now()
	if requestBody.PublishDate != nil {
		publishDate = *requestBody.PublishDate
	}

	// Create news object
	news := models.News{
		Title:       requestBody.Title,
		Slug:        newsSlug,
		Content:     requestBody.Content,
		Summary:     requestBody.Summary,
		Source:      requestBody.Source,
		SourceURL:   requestBody.SourceURL,
		ImageURL:    requestBody.ImageURL,
		Category:    requestBody.Category,
		Status:      requestBody.Status,
		Published:   requestBody.Status == models.NewsStatusPublished,
		PublishDate: publishDate,
	}

	// If no category is provided, use technology
	if news.Category == "" {
		news.Category = models.NewsCategoryTechnology
	}

	// If no status is provided, use draft
	if news.Status == "" {
		news.Status = models.NewsStatusDraft
		news.Published = false
	}

	// Begin transaction
	tx := database.DB.Begin()

	// Create news article
	if err := tx.Create(&news).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("Failed to create news article")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create news article"})
		return
	}

	// Add tags if provided
	if len(requestBody.Tags) > 0 {
		for _, tagName := range requestBody.Tags {
			var tag models.Tag
			tagName = strings.TrimSpace(tagName)

			// Find or create tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				log.Error().Err(err).Str("tag", tagName).Msg("Failed to process tags")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tags"})
				return
			}

			// Associate tag with news
			if err := tx.Model(&news).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				log.Error().Err(err).Str("tag", tagName).Msg("Failed to associate tags")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
				return
			}
		}
	}

	// Commit transaction
	tx.Commit()

	// Reload news with tags
	database.DB.Preload("Tags").First(&news, news.ID)

	c.JSON(http.StatusCreated, news)
}

// UpdateNews godoc
// @Summary Update a news article
// @Description Update an existing news article (admin only)
// @Tags News
// @Accept json
// @Produce json
// @Param id path int true "News article ID"
// @Param news body models.UpdateNewsRequest true "Updated news article"
// @Success 200 {object} models.News "Updated news article"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news/{id} [put]
func UpdateNews(c *gin.Context) {
	// Get news ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	// Find existing news
	var news models.News
	if err := database.DB.Preload("Tags").First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Uint64("id", id).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news article"})
		}
		return
	}

	// Parse request body
	var requestBody models.UpdateNewsRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update news fields if provided
	if requestBody.Title != "" {
		// If title is changing, update slug
		if news.Title != requestBody.Title {
			newSlug := slug.Make(requestBody.Title)

			// Check if new slug already exists
			var count int64
			if err := database.DB.Model(&models.News{}).Where("slug = ? AND id != ?", newSlug, id).Count(&count).Error; err != nil {
				log.Error().Err(err).Str("slug", newSlug).Msg("Failed to check for existing slug")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news article"})
				return
			}

			// If slug exists, append a timestamp
			if count > 0 {
				newSlug = fmt.Sprintf("%s-%d", newSlug, time.Now().Unix())
			}

			news.Slug = newSlug
		}
		news.Title = requestBody.Title
	}

	if requestBody.Content != "" {
		news.Content = requestBody.Content
	}

	if requestBody.Summary != "" {
		news.Summary = requestBody.Summary
	}

	if requestBody.Source != "" {
		news.Source = requestBody.Source
	}

	if requestBody.SourceURL != "" {
		news.SourceURL = requestBody.SourceURL
	}

	if requestBody.ImageURL != "" {
		news.ImageURL = requestBody.ImageURL
	}

	if requestBody.Category != "" {
		news.Category = requestBody.Category
	}

	if requestBody.Status != "" {
		news.Status = requestBody.Status
		news.Published = requestBody.Status == models.NewsStatusPublished
	}

	if requestBody.PublishDate != nil {
		news.PublishDate = *requestBody.PublishDate
	}

	// Begin transaction
	tx := database.DB.Begin()

	// Update news
	if err := tx.Save(&news).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Uint64("id", id).Msg("Failed to update news article")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news article"})
		return
	}

	// Update tags if provided
	if len(requestBody.Tags) > 0 {
		// Clear existing tags
		if err := tx.Model(&news).Association("Tags").Clear(); err != nil {
			tx.Rollback()
			log.Error().Err(err).Uint64("id", id).Msg("Failed to clear existing tags")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags"})
			return
		}

		// Add new tags
		for _, tagName := range requestBody.Tags {
			var tag models.Tag
			tagName = strings.TrimSpace(tagName)

			// Find or create tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				log.Error().Err(err).Str("tag", tagName).Msg("Failed to process tags")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process tags"})
				return
			}

			// Associate tag with news
			if err := tx.Model(&news).Association("Tags").Append(&tag); err != nil {
				tx.Rollback()
				log.Error().Err(err).Str("tag", tagName).Msg("Failed to associate tags")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
				return
			}
		}
	}

	// Commit transaction
	tx.Commit()

	// Reload news with tags
	database.DB.Preload("Tags").First(&news, news.ID)

	c.JSON(http.StatusOK, news)
}

// DeleteNews godoc
// @Summary Delete a news article
// @Description Delete a news article (admin only)
// @Tags News
// @Produce json
// @Param id path int true "News article ID"
// @Success 200 {object} models.SwaggerDeleteNewsResponse "News article deleted"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news/{id} [delete]
func DeleteNews(c *gin.Context) {
	// Get news ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	// Check if news exists
	var news models.News
	if err := database.DB.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Uint64("id", id).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete news article"})
		}
		return
	}

	// Begin transaction
	tx := database.DB.Begin()

	// Clear associations
	if err := tx.Model(&news).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		log.Error().Err(err).Uint64("id", id).Msg("Failed to clear tags")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete news article"})
		return
	}

	// Delete news
	if err := tx.Delete(&news).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Uint64("id", id).Msg("Failed to delete news article")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete news article"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "News article deleted successfully"})
}

// SetNewsStatus godoc
// @Summary Set news article status
// @Description Update the status of a news article (admin only)
// @Tags News
// @Accept json
// @Produce json
// @Param id path int true "News article ID"
// @Param status body models.SetNewsStatusRequest true "New status"
// @Success 200 {object} models.News "Updated news article"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news/{id}/status [post]
func SetNewsStatus(c *gin.Context) {
	// Get news ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	// Parse request body
	var requestBody models.SetNewsStatusRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Find news
	var news models.News
	if err := database.DB.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Uint64("id", id).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news status"})
		}
		return
	}

	// Update status
	news.Status = requestBody.Status
	news.Published = requestBody.Status == models.NewsStatusPublished
	news.UpdatedAt = time.Now()

	// If setting to published, set publish date to now if not already set
	if news.Status == models.NewsStatusPublished && news.PublishDate.Before(time.Now().AddDate(0, 0, -1)) {
		news.PublishDate = time.Now()
	}

	// Save changes
	if err := database.DB.Save(&news).Error; err != nil {
		log.Error().Err(err).Uint64("id", id).Msg("Failed to update news status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news status"})
		return
	}

	// Reload news with tags
	database.DB.Preload("Tags").First(&news, news.ID)

	c.JSON(http.StatusOK, news)
}

// FetchExternalNews godoc
// @Summary Fetch news from external API
// @Description Fetch and store news from external API (admin only)
// @Tags News
// @Accept json
// @Produce json
// @Param fetch_request body models.FetchNewsRequest true "Fetch request parameters"
// @Success 200 {object} models.SwaggerFetchNewsResponse "News articles fetched"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news/fetch [post]
func FetchExternalNews(c *gin.Context) {
	// Parse request body
	var requestBody models.FetchNewsRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Initialize News API service
	newsService, err := services.NewNewsService(middleware.AppConfig.NewsAPI)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize NewsAPI service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize news service"})
		return
	}

	// Fetch news
	news, err := newsService.FetchNews(c.Request.Context(), requestBody.Categories, requestBody.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch news")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news from external API"})
		return
	}

	if len(news) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No news articles found"})
		return
	}

	// Store each news article in separate transactions to avoid
	// aborting the entire batch if one fails
	var savedCount int
	for _, article := range news {
		// Start a new transaction for each article
		tx := database.DB.Begin()

		// Check if article already exists by external ID
		var existingCount int64
		if err := tx.Model(&models.News{}).Where("external_id = ?", article.ExternalID).Count(&existingCount).Error; err != nil {
			log.Error().Err(err).Str("external_id", article.ExternalID).Msg("Failed to check existing news")
			tx.Rollback()
			continue
		}

		if existingCount > 0 {
			tx.Rollback() // Clean rollback for skipped articles
			continue      // Skip existing articles
		}

		// Check if slug already exists
		var slugCount int64
		if err := tx.Model(&models.News{}).Where("slug = ?", article.Slug).Count(&slugCount).Error; err != nil {
			log.Error().Err(err).Str("slug", article.Slug).Msg("Failed to check existing slug")
			tx.Rollback()
			continue
		}

		// If slug exists, make it unique by adding a timestamp
		if slugCount > 0 {
			article.Slug = fmt.Sprintf("%s-%d", article.Slug, time.Now().Unix())
		}

		// Save article
		if err := tx.Create(&article).Error; err != nil {
			log.Error().Err(err).Str("title", article.Title).Msg("Failed to save news article")
			tx.Rollback()
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			log.Error().Err(err).Str("title", article.Title).Msg("Failed to commit transaction")
			continue
		}

		savedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "News articles fetched successfully",
		"total":      len(news),
		"saved":      savedCount,
		"categories": requestBody.Categories,
		"fetch_time": time.Now().Format(time.RFC3339),
	})
}

// FetchRSSNews godoc
// @Summary Fetch news from RSS feeds
// @Description Fetch and store news from configured RSS feeds (admin only)
// @Tags News
// @Accept json
// @Produce json
// @Param fetch_request body models.FetchNewsRequest true "Fetch request parameters (only limit is used for RSS feeds)"
// @Success 200 {object} models.SwaggerFetchRSSNewsResponse "News articles fetched from RSS feeds"
// @Failure 400 {object} models.SwaggerStandardResponse "Invalid input"
// @Failure 401 {object} models.SwaggerStandardResponse "Unauthorized"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Security BearerAuth
// @Router /admin/news/fetch-rss [post]
func FetchRSSNews(c *gin.Context) {
	// Parse request body
	var requestBody models.FetchNewsRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Initialize RSS service
	rssService, err := services.NewRSSService(middleware.AppConfig.RSS)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize RSS service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize RSS service"})
		return
	}

	// Fetch news from RSS feeds
	news, err := rssService.FetchNews(c.Request.Context(), requestBody.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch news from RSS feeds")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news from RSS feeds"})
		return
	}

	if len(news) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No news articles found in RSS feeds"})
		return
	}

	// Store each news article in separate transactions to avoid
	// aborting the entire batch if one fails
	var savedCount int
	for _, article := range news {
		// Start a new transaction for each article
		tx := database.DB.Begin()

		// Check if article already exists by external ID
		var existingCount int64
		if err := tx.Model(&models.News{}).Where("external_id = ?", article.ExternalID).Count(&existingCount).Error; err != nil {
			log.Error().Err(err).Str("external_id", article.ExternalID).Msg("Failed to check existing news")
			tx.Rollback()
			continue
		}

		if existingCount > 0 {
			tx.Rollback() // Clean rollback for skipped articles
			continue      // Skip existing articles
		}

		// Check if slug already exists
		var slugCount int64
		if err := tx.Model(&models.News{}).Where("slug = ?", article.Slug).Count(&slugCount).Error; err != nil {
			log.Error().Err(err).Str("slug", article.Slug).Msg("Failed to check existing slug")
			tx.Rollback()
			continue
		}

		// If slug exists, make it unique by adding a timestamp
		if slugCount > 0 {
			article.Slug = fmt.Sprintf("%s-%d", article.Slug, time.Now().Unix())
		}

		// Save article
		if err := tx.Create(&article).Error; err != nil {
			log.Error().Err(err).Str("title", article.Title).Msg("Failed to save news article")
			tx.Rollback()
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			log.Error().Err(err).Str("title", article.Title).Msg("Failed to commit transaction")
			continue
		}

		savedCount++
	}

	// Collect all unique categories from the fetched news
	categories := make(map[models.NewsCategory]bool)
	for _, article := range news {
		categories[article.Category] = true
	}

	// Convert categories map to slice for response
	var categoriesSlice []models.NewsCategory
	for category := range categories {
		categoriesSlice = append(categoriesSlice, category)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "RSS news articles fetched successfully",
		"total":      len(news),
		"saved":      savedCount,
		"categories": categoriesSlice,
		"fetch_time": time.Now().Format(time.RFC3339),
	})
}

// GetNewsFullContent godoc
// @Summary Get full content for news article
// @Description Attempts to fetch and return the full content for a news article
// @Tags News
// @Produce json
// @Param id path int true "News article ID"
// @Success 200 {object} models.SwaggerNewsWithContentStatus "News article with full content status"
// @Failure 404 {object} models.SwaggerStandardResponse "News article not found"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /news/{id}/full-content [get]
func GetNewsFullContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	// Get the news article
	var news models.News
	if err := database.DB.
		Where("id = ? AND status = ? AND published = ?", id, models.NewsStatusPublished, true).
		Preload("Tags").
		First(&news).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "News article not found"})
		} else {
			log.Error().Err(err).Str("id", id).Msg("Failed to retrieve news article")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news article"})
		}
		return
	}

	// Check if we already have enriched content in the database
	var enrichedContent models.EnrichedNewsContent
	enrichedContentExists := true

	if err := database.DB.Where("news_id = ?", news.ID).First(&enrichedContent).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Error().Err(err).Uint("newsID", news.ID).Msg("Failed to check for enriched content")
		}
		enrichedContentExists = false
	}

	// Prepare content status
	contentStatus := models.ContentStatus{
		IsTruncated:    services.IsTruncated(news.Content),
		TruncatedChars: 0,
		HasFullContent: false,
		FetchError:     "",
	}

	// Update content status based on existing enriched content
	if enrichedContentExists {
		contentStatus.IsTruncated = enrichedContent.IsTruncated
		contentStatus.TruncatedChars = enrichedContent.TruncatedChars
		contentStatus.HasFullContent = enrichedContent.FullContent != ""
		contentStatus.FetchError = enrichedContent.FetchError

		// If the enriched content is recent (less than 24 hours old), use it
		if !enrichedContent.LastFetched.IsZero() && time.Since(enrichedContent.LastFetched) < 24*time.Hour {
			// If we have full content, use it instead of the original
			if enrichedContent.FullContent != "" {
				news.Content = enrichedContent.FullContent
			}

			c.JSON(http.StatusOK, models.NewsWithContentStatus{
				News:          news,
				ContentStatus: contentStatus,
			})
			return
		}
	}

	// If we don't have recent enriched content or it doesn't exist,
	// attempt to fetch it now
	contentScraper := services.NewContentScraper()
	enriched, err := contentScraper.EnrichNewsContent(c.Request.Context(), &news)
	if err != nil {
		log.Error().Err(err).Uint("newsID", news.ID).Msg("Failed to enrich news content")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve full content"})
		return
	}

	// Update or create the enriched content record
	if enrichedContentExists {
		// Update existing record
		enrichedContent.FullContent = enriched.FullContent
		enrichedContent.IsTruncated = enriched.IsTruncated
		enrichedContent.TruncatedChars = enriched.TruncatedChars
		enrichedContent.TruncationPattern = enriched.TruncationPattern
		enrichedContent.LastFetched = time.Now()
		enrichedContent.FetchError = enriched.FetchError
		enrichedContent.UpdatedAt = time.Now()

		if err := database.DB.Save(&enrichedContent).Error; err != nil {
			log.Error().Err(err).Uint("newsID", news.ID).Msg("Failed to update enriched content")
		}
	} else {
		// Create new record
		enrichedContent = models.EnrichedNewsContent{
			NewsID:             news.ID,
			OriginalContent:    news.Content,
			FullContent:        enriched.FullContent,
			IsTruncated:        enriched.IsTruncated,
			TruncatedChars:     enriched.TruncatedChars,
			TruncationPattern:  enriched.TruncationPattern,
			SourceURL:          news.SourceURL,
			LastFetched:        time.Now(),
			TruncationDetected: time.Now(),
			FetchError:         enriched.FetchError,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := database.DB.Create(&enrichedContent).Error; err != nil {
			log.Error().Err(err).Uint("newsID", news.ID).Msg("Failed to save enriched content")
		}
	}

	// Update content status
	contentStatus.IsTruncated = enriched.IsTruncated
	contentStatus.TruncatedChars = enriched.TruncatedChars
	contentStatus.HasFullContent = enriched.FullContent != ""
	contentStatus.FetchError = enriched.FetchError

	// If we have full content, use it instead of the original
	if enriched.FullContent != "" {
		news.Content = enriched.FullContent
	}

	c.JSON(http.StatusOK, models.NewsWithContentStatus{
		News:          news,
		ContentStatus: contentStatus,
	})
}
