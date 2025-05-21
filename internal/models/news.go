package models

import (
	"time"

	"gorm.io/gorm"
)

// NewsStatus represents the publication status of a news article
type NewsStatus string

const (
	// NewsStatusPublished indicates the news article is published and publicly visible
	NewsStatusPublished NewsStatus = "published"
	// NewsStatusDraft indicates the news article is a draft and not publicly visible
	NewsStatusDraft NewsStatus = "draft"
	// NewsStatusArchived indicates the news article has been archived
	NewsStatusArchived NewsStatus = "archived"
)

// NewsCategory represents the category of a news article
type NewsCategory string

const (
	// NewsCategoryGeneral represents general news
	NewsCategoryGeneral NewsCategory = "general"
	// NewsCategoryBusiness represents business news
	NewsCategoryBusiness NewsCategory = "business"
	// NewsCategoryTechnology represents technology news
	NewsCategoryTechnology NewsCategory = "technology"
	// NewsCategoryScience represents science news
	NewsCategoryScience NewsCategory = "science"
	// NewsCategoryHealth represents health news
	NewsCategoryHealth NewsCategory = "health"
	// NewsCategorySports represents sports news
	NewsCategorySports NewsCategory = "sports"
	// NewsCategoryEntertainment represents entertainment news
	NewsCategoryEntertainment NewsCategory = "entertainment"
)

// News represents a news article
// @Description A news article with content, metadata, and relationships
type News struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Title       string         `json:"title" gorm:"size:255;not null" example:"Major Technology Breakthrough Announced" description:"News title"`
	Slug        string         `json:"slug" gorm:"size:255;not null;unique" example:"major-technology-breakthrough-announced" description:"URL-friendly version of the title"`
	Content     string         `json:"content" gorm:"type:text;not null" example:"Scientists announced a major breakthrough in quantum computing..." description:"Main content of the news article"`
	Summary     string         `json:"summary" gorm:"type:text" example:"A brief summary of the quantum computing breakthrough" description:"Short summary of the news article"`
	Source      string         `json:"source" gorm:"size:100;not null" example:"TechNews" description:"Original source of the news"`
	SourceURL   string         `json:"source_url" gorm:"size:500" example:"https://technews.com/article/12345" description:"URL to the original news article"`
	ImageURL    string         `json:"image_url" gorm:"size:500" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/news/article1.jpg" description:"URL to the news article's image"`
	Category    NewsCategory   `json:"category" gorm:"type:varchar(20);not null;default:'general'" example:"technology" description:"Category of the news article"`
	Status      NewsStatus     `json:"status" gorm:"type:varchar(20);not null;default:'published'" example:"published" description:"Publication status of the news article"`
	Published   bool           `json:"published" gorm:"default:true" example:"true" description:"Whether the news is published and visible"`
	PublishDate time.Time      `json:"publish_date" example:"2023-01-01T12:00:00Z" description:"When the news was/will be published"`
	ExternalID  string         `json:"external_id" gorm:"size:100;index" example:"ext-12345" description:"ID from external news API"`
	CreatedAt   time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the news article was created"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the news article was last updated"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
	Tags        []Tag          `json:"tags" gorm:"many2many:news_tags;" description:"Tags associated with the news article"`
}

// CreateNewsRequest represents the request body for creating a news article
// @Description Request model for creating a news article
type CreateNewsRequest struct {
	Title       string       `json:"title" binding:"required" example:"Major Technology Breakthrough Announced" description:"News title"`
	Content     string       `json:"content" binding:"required" example:"Scientists announced a major breakthrough in quantum computing..." description:"Main content of the news article"`
	Summary     string       `json:"summary" example:"A brief summary of the quantum computing breakthrough" description:"Short summary of the news article"`
	Source      string       `json:"source" binding:"required" example:"TechNews" description:"Original source of the news"`
	SourceURL   string       `json:"source_url" example:"https://technews.com/article/12345" description:"URL to the original news article"`
	ImageURL    string       `json:"image_url" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/news/article1.jpg" description:"URL to the news article's image"`
	Category    NewsCategory `json:"category" example:"technology" description:"Category of the news article"`
	Status      NewsStatus   `json:"status" example:"published" description:"Publication status of the news article"`
	PublishDate *time.Time   `json:"publish_date" example:"2023-01-01T12:00:00Z" description:"When the news will be published"`
	Tags        []string     `json:"tags" example:"['technology', 'quantum computing']" description:"Tags to associate with the news article"`
}

// UpdateNewsRequest represents the request body for updating a news article
// @Description Request model for updating a news article
type UpdateNewsRequest struct {
	Title       string       `json:"title" example:"Updated Technology Breakthrough Announced" description:"News title"`
	Content     string       `json:"content" example:"Scientists announced a major breakthrough in quantum computing..." description:"Main content of the news article"`
	Summary     string       `json:"summary" example:"A brief summary of the quantum computing breakthrough" description:"Short summary of the news article"`
	Source      string       `json:"source" example:"TechNews" description:"Original source of the news"`
	SourceURL   string       `json:"source_url" example:"https://technews.com/article/12345" description:"URL to the original news article"`
	ImageURL    string       `json:"image_url" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/news/article1.jpg" description:"URL to the news article's image"`
	Category    NewsCategory `json:"category" example:"technology" description:"Category of the news article"`
	Status      NewsStatus   `json:"status" example:"published" description:"Publication status of the news article"`
	PublishDate *time.Time   `json:"publish_date" example:"2023-01-01T12:00:00Z" description:"When the news will be published"`
	Tags        []string     `json:"tags" example:"['technology', 'quantum computing']" description:"Tags to associate with the news article"`
}

// NewsResponse represents a news response with pagination
// @Description Response model for news list with pagination information
type NewsResponse struct {
	News       []News `json:"news" description:"List of news articles"`
	TotalItems int64  `json:"total_items" example:"100" description:"Total number of news articles"`
	Page       int    `json:"page" example:"1" description:"Current page number"`
	PerPage    int    `json:"per_page" example:"10" description:"Number of items per page"`
	TotalPages int    `json:"total_pages" example:"10" description:"Total number of pages"`
}

// NewsQuery represents query parameters for filtering news articles
// @Description Query parameters for filtering news articles
type NewsQuery struct {
	Category string `form:"category" json:"category" example:"technology" description:"Filter by category"`
	Tag      string `form:"tag" json:"tag" example:"technology" description:"Filter by tag"`
	Search   string `form:"search" json:"search" example:"quantum" description:"Search in title and content"`
	Page     int    `form:"page" json:"page" example:"1" description:"Page number"`
	PerPage  int    `form:"per_page" json:"per_page" example:"10" description:"Items per page"`
}

// SetNewsStatusRequest represents the request body for updating a news article's status
// @Description Request model for changing a news article's status
type SetNewsStatusRequest struct {
	Status NewsStatus `json:"status" binding:"required" example:"published" description:"New news status (published, draft, archived)"`
}

// FetchNewsRequest represents the request body for fetching news from external API
// @Description Request model for fetching news from external API
type FetchNewsRequest struct {
	Categories []NewsCategory `json:"categories" example:"['technology', 'business']" description:"Categories of news to fetch"`
	Limit      int            `json:"limit" example:"10" description:"Maximum number of news articles to fetch"`
}
