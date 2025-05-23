package models

import "time"

// SwaggerFetchNewsResponse represents the response for fetching news from external API
// @Description Response format for fetching news from external API
type SwaggerFetchNewsResponse struct {
	Message    string         `json:"message" example:"News articles fetched successfully" description:"Status message"`
	Total      int            `json:"total" example:"10" description:"Total number of news articles fetched"`
	Saved      int            `json:"saved" example:"8" description:"Number of new articles saved"`
	Categories []NewsCategory `json:"categories" example:"[\"technology\",\"business\"]" description:"Categories that were fetched"`
	FetchTime  string         `json:"fetch_time" example:"2023-01-01T12:00:00Z" description:"When the fetch operation was performed"`
}

// SwaggerFetchRSSNewsResponse represents the response for fetching news from RSS feeds
// @Description Response format for fetching news from RSS feeds
type SwaggerFetchRSSNewsResponse struct {
	Message    string         `json:"message" example:"RSS news articles fetched successfully" description:"Status message"`
	Total      int            `json:"total" example:"15" description:"Total number of news articles fetched from RSS feeds"`
	Saved      int            `json:"saved" example:"12" description:"Number of new articles saved"`
	Categories []NewsCategory `json:"categories" example:"[\"technology\",\"science\"]" description:"Categories of the fetched articles"`
	FetchTime  string         `json:"fetch_time" example:"2023-01-01T12:00:00Z" description:"When the fetch operation was performed"`
}

// SwaggerDeleteNewsResponse represents the response for deleting a news article
// @Description Response format for deleting a news article
type SwaggerDeleteNewsResponse struct {
	Message string `json:"message" example:"News article deleted successfully" description:"Status message"`
}

// SwaggerNewsWithContentStatus represents a news article with content status information
// @Description News article with content status information for Swagger documentation
type SwaggerNewsWithContentStatus struct {
	News          News          `json:"news" description:"The news article"`
	ContentStatus ContentStatus `json:"content_status" description:"Status of the article content"`
}

// SwaggerEnrichedNewsContent represents the response for enriched news content
// @Description Enriched news content with full article text
type SwaggerEnrichedNewsContent struct {
	NewsID             uint      `json:"news_id" example:"1" description:"ID of the associated news article"`
	OriginalContent    string    `json:"original_content" example:"Truncated content..." description:"Original content from the news API"`
	FullContent        string    `json:"full_content" example:"Full article content retrieved from the source..." description:"Full content fetched from the source"`
	IsTruncated        bool      `json:"is_truncated" example:"true" description:"Whether the original content was truncated"`
	TruncatedChars     int       `json:"truncated_chars" example:"1281" description:"Number of characters truncated if known"`
	TruncationPattern  string    `json:"truncation_pattern" example:"[+1281 chars]" description:"The pattern indicating truncation"`
	SourceURL          string    `json:"source_url" example:"https://news.com/article" description:"URL where full content was fetched from"`
	LastFetched        time.Time `json:"last_fetched" example:"2023-01-01T12:00:00Z" description:"When the full content was last fetched"`
	TruncationDetected time.Time `json:"truncation_detected" example:"2023-01-01T12:00:00Z" description:"When truncation was first detected"`
	FetchError         string    `json:"fetch_error" example:"" description:"Error message if fetch failed"`
}

// SwaggerContentStatus represents the content status for swagger documentation
// @Description Status information about the news content
type SwaggerContentStatus struct {
	IsTruncated    bool   `json:"is_truncated" example:"true" description:"Whether the content is truncated"`
	TruncatedChars int    `json:"truncated_chars" example:"1281" description:"Number of characters truncated if known"`
	HasFullContent bool   `json:"has_full_content" example:"true" description:"Whether full content is available"`
	FetchError     string `json:"fetch_error,omitempty" example:"" description:"Error message if fetch failed"`
}

// SwaggerNewsWithoutContent represents a news article without content field
// @Description A news article without the content field for improved performance
type SwaggerNewsWithoutContent struct {
	ID          uint         `json:"id" example:"1" description:"Unique identifier"`
	Title       string       `json:"title" example:"Major Technology Breakthrough Announced" description:"News title"`
	Slug        string       `json:"slug" example:"major-technology-breakthrough-announced" description:"URL-friendly version of the title"`
	Summary     string       `json:"summary" example:"A brief summary of the quantum computing breakthrough" description:"Short summary of the news article"`
	Source      string       `json:"source" example:"TechNews" description:"Original source of the news"`
	SourceURL   string       `json:"source_url" example:"https://technews.com/article/12345" description:"URL to the original news article"`
	ImageURL    string       `json:"image_url" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/news/article1.jpg" description:"URL to the news article's image"`
	Category    NewsCategory `json:"category" example:"technology" description:"Category of the news article"`
	Status      NewsStatus   `json:"status" example:"published" description:"Publication status of the news article"`
	Published   bool         `json:"published" example:"true" description:"Whether the news is published and visible"`
	PublishDate time.Time    `json:"publish_date" example:"2023-01-01T12:00:00Z" description:"When the news was/will be published"`
	CreatedAt   time.Time    `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the news article was created"`
	UpdatedAt   time.Time    `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the news article was last updated"`
	Tags        []Tag        `json:"tags" description:"Tags associated with the news article"`
}

// SwaggerNewsWithoutContentResponse represents a response with news articles without content
// @Description Response model for news list with pagination and without content fields
type SwaggerNewsWithoutContentResponse struct {
	News       []SwaggerNewsWithoutContent `json:"news" description:"List of news articles without content"`
	TotalItems int64                       `json:"total_items" example:"100" description:"Total number of news articles"`
	Page       int                         `json:"page" example:"1" description:"Current page number"`
	PerPage    int                         `json:"per_page" example:"10" description:"Number of items per page"`
	TotalPages int                         `json:"total_pages" example:"10" description:"Total number of pages"`
}
