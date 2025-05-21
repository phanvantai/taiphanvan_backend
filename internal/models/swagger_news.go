package models

// SwaggerFetchNewsResponse represents the response for fetching news from external API
// @Description Response format for fetching news from external API
type SwaggerFetchNewsResponse struct {
	Message    string         `json:"message" example:"News articles fetched successfully" description:"Status message"`
	Total      int            `json:"total" example:"10" description:"Total number of news articles fetched"`
	Saved      int            `json:"saved" example:"8" description:"Number of new articles saved"`
	Categories []NewsCategory `json:"categories" example:"[\"technology\",\"business\"]" description:"Categories that were fetched"`
	FetchTime  string         `json:"fetch_time" example:"2023-01-01T12:00:00Z" description:"When the fetch operation was performed"`
}

// SwaggerDeleteNewsResponse represents the response for deleting a news article
// @Description Response format for deleting a news article
type SwaggerDeleteNewsResponse struct {
	Message string `json:"message" example:"News article deleted successfully" description:"Status message"`
}
