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
