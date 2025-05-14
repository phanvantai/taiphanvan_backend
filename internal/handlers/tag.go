package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
)

// fetchTags is a generic helper function to fetch tags with post counts
func fetchTags[T any](orderBy string, limit int) ([]T, error) {
	var result []T
	query := database.DB.Table("tags").
		Select("tags.id, tags.name, COUNT(DISTINCT post_tags.post_id) as post_count").
		Joins("LEFT JOIN post_tags ON post_tags.tag_id = tags.id").
		Group("tags.id").
		Order(orderBy)

	if limit > 0 {
		query = query.Limit(limit)
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item T
		if err := database.DB.ScanRows(rows, &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

// GetAllTags godoc
// @Summary Get all tags
// @Description Returns all tags with their post counts
// @Tags Tags
// @Produce json
// @Success 200 {object} models.StandardResponse "List of tags with post counts"
// @Failure 500 {object} models.StandardResponse "Server error"
// @Router /tags [get]
func GetAllTags(c *gin.Context) {
	tags, err := fetchTags[models.TagWithCount]("tags.name", 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to fetch tags"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(tags, "Tags retrieved successfully"))
}

// GetPopularTags godoc
// @Summary Get popular tags
// @Description Returns the most used tags with post counts (limited to 10)
// @Tags Tags
// @Produce json
// @Success 200 {object} models.StandardResponse "List of popular tags with post counts"
// @Failure 500 {object} models.StandardResponse "Server error"
// @Router /tags/popular [get]
func GetPopularTags(c *gin.Context) {
	limit := 10 // Default limit

	tags, err := fetchTags[models.TagWithCount]("post_count DESC, tags.name", limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to fetch popular tags"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(tags, "Popular tags retrieved successfully"))
}
