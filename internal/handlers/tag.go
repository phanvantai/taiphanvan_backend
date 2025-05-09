package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
)

// GetAllTags godoc
// @Summary Get all tags
// @Description Returns all tags with their post counts
// @Tags Tags
// @Produce json
// @Success 200 {array} models.TagWithCount "List of tags with post counts"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /tags [get]
func GetAllTags(c *gin.Context) {
	var tagsWithCount []models.TagWithCount

	rows, err := database.DB.Table("tags").
		Select("tags.id, tags.name, COUNT(DISTINCT post_tags.post_id) as post_count").
		Joins("LEFT JOIN post_tags ON post_tags.tag_id = tags.id").
		Group("tags.id").
		Order("tags.name").
		Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tag models.TagWithCount
		if err := database.DB.ScanRows(rows, &tag); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan tags"})
			return
		}
		tagsWithCount = append(tagsWithCount, tag)
	}

	c.JSON(http.StatusOK, tagsWithCount)
}

// GetPopularTags godoc
// @Summary Get popular tags
// @Description Returns the most used tags with post counts (limited to 10)
// @Tags Tags
// @Produce json
// @Success 200 {array} models.TagWithCount "List of popular tags with post counts"
// @Failure 500 {object} models.SwaggerStandardResponse "Server error"
// @Router /tags/popular [get]
func GetPopularTags(c *gin.Context) {
	limit := 10 // Default limit

	type TagWithCount struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		PostCount int64  `json:"post_count"`
	}

	var tagsWithCount []TagWithCount

	rows, err := database.DB.Table("tags").
		Select("tags.id, tags.name, COUNT(DISTINCT post_tags.post_id) as post_count").
		Joins("LEFT JOIN post_tags ON post_tags.tag_id = tags.id").
		Group("tags.id").
		Order("post_count DESC, tags.name").
		Limit(limit).
		Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular tags"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tag TagWithCount
		if err := database.DB.ScanRows(rows, &tag); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan tags"})
			return
		}
		tagsWithCount = append(tagsWithCount, tag)
	}

	c.JSON(http.StatusOK, tagsWithCount)
}
