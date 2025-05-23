package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
)

// NewsService handles interactions with external news APIs
type NewsService struct {
	cfg        config.NewsAPIConfig
	httpClient *http.Client
}

// NewsAPIResponse represents the response from the NewsAPI
type NewsAPIResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source"`
		Author      string    `json:"author"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		URL         string    `json:"url"`
		URLToImage  string    `json:"urlToImage"`
		PublishedAt time.Time `json:"publishedAt"`
		Content     string    `json:"content"`
	} `json:"articles"`
}

// NewNewsService creates a new NewsAPI service
func NewNewsService(cfg config.NewsAPIConfig) (*NewsService, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("missing NewsAPI API key")
	}

	return &NewsService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// FetchNews fetches news articles from the NewsAPI
func (s *NewsService) FetchNews(ctx context.Context, categories []models.NewsCategory, limit int) ([]models.News, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	maxLimit := 100
	if limit > maxLimit {
		limit = maxLimit // Cap at maximum
	}

	var allNews []models.News

	// If no categories specified, use all available categories
	if len(categories) == 0 {
		categories = []models.NewsCategory{
			//models.NewsCategoryGeneral,
			//models.NewsCategoryBusiness,
			models.NewsCategoryTechnology,
			models.NewsCategoryScience,
			// models.NewsCategoryHealth,
			// models.NewsCategorySports,
			// models.NewsCategoryEntertainment,
		}
	}

	// Fetch news for each category
	for _, category := range categories {
		categoryNews, err := s.fetchNewsByCategory(ctx, string(category), limit/len(categories))
		if err != nil {
			log.Error().Err(err).Str("category", string(category)).Msg("Failed to fetch news for category")
			continue // Continue with other categories
		}

		allNews = append(allNews, categoryNews...)
	}

	return allNews, nil
}

// fetchNewsByCategory fetches news articles by category
func (s *NewsService) fetchNewsByCategory(ctx context.Context, category string, limit int) ([]models.News, error) {
	// Build URL
	apiURL, err := url.Parse(s.cfg.BaseURL + "/top-headlines")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("apiKey", s.cfg.APIKey)
	params.Add("category", category)
	params.Add("pageSize", fmt.Sprintf("%d", limit))
	params.Add("language", "en") // English language
	apiURL.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	log.Info().Str("url", apiURL.String()).Str("category", category).Msg("Fetching news from NewsAPI")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var newsResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check response status
	if newsResp.Status != "ok" {
		return nil, fmt.Errorf("API returned non-OK status: %s", newsResp.Status)
	}

	// Convert to News models
	var news []models.News
	newsCategory := models.NewsCategory(category)

	for _, article := range newsResp.Articles {
		// Generate a unique slug
		titleSlug := slug.Make(article.Title)

		// Create a more compact external ID that won't exceed 100 chars
		// Use the source ID (or "unknown" if none) and the last part of the URL
		sourceId := article.Source.ID
		if sourceId == "" {
			sourceId = "unknown"
		}

		// Extract domain from URL to keep the external ID shorter
		parsedURL, err := url.Parse(article.URL)
		urlPart := ""
		if err == nil {
			urlPart = parsedURL.Host + parsedURL.Path
		} else {
			urlPart = article.URL
		}

		// Ensure the external ID doesn't exceed 100 characters
		externalID := fmt.Sprintf("%s-%s", sourceId, urlPart)
		if len(externalID) > 100 {
			externalID = externalID[:100]
		}

		// Create the news article
		newsArticle := models.News{
			Title:       article.Title,
			Slug:        titleSlug,
			Content:     article.Content,
			Summary:     article.Description,
			Source:      article.Source.Name,
			SourceURL:   article.URL,
			ImageURL:    article.URLToImage,
			Category:    newsCategory,
			Status:      models.NewsStatusPublished,
			Published:   true,
			PublishDate: article.PublishedAt,
			ExternalID:  externalID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		news = append(news, newsArticle)
	}

	log.Info().Int("count", len(news)).Str("category", category).Msg("Successfully fetched news articles")
	return news, nil
}

// SearchNews searches for news articles from the NewsAPI
func (s *NewsService) SearchNews(ctx context.Context, query string, limit int) ([]models.News, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	maxLimit := 100
	if limit > maxLimit {
		limit = maxLimit // Cap at maximum
	}

	// Build URL
	apiURL, err := url.Parse(s.cfg.BaseURL + "/everything")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("apiKey", s.cfg.APIKey)
	params.Add("q", query)
	params.Add("pageSize", fmt.Sprintf("%d", limit))
	params.Add("language", "en") // English language
	params.Add("sortBy", "relevancy")
	apiURL.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	log.Info().Str("url", apiURL.String()).Str("query", query).Msg("Searching news from NewsAPI")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var newsResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check response status
	if newsResp.Status != "ok" {
		return nil, fmt.Errorf("API returned non-OK status: %s", newsResp.Status)
	}

	// Convert to News models
	var news []models.News

	for _, article := range newsResp.Articles {
		// Determine category based on content analysis
		category := determineCategory(article.Title, article.Description)

		// Generate a unique slug
		titleSlug := slug.Make(article.Title)

		// Create a more compact external ID that won't exceed 100 chars
		sourceId := article.Source.ID
		if sourceId == "" {
			sourceId = "unknown"
		}

		// Extract domain from URL to keep the external ID shorter
		parsedURL, err := url.Parse(article.URL)
		urlPart := ""
		if err == nil {
			urlPart = parsedURL.Host + parsedURL.Path
		} else {
			urlPart = article.URL
		}

		// Ensure the external ID doesn't exceed 100 characters
		externalID := fmt.Sprintf("%s-%s", sourceId, urlPart)
		if len(externalID) > 100 {
			externalID = externalID[:100]
		}

		// Create the news article
		newsArticle := models.News{
			Title:       article.Title,
			Slug:        titleSlug,
			Content:     article.Content,
			Summary:     article.Description,
			Source:      article.Source.Name,
			SourceURL:   article.URL,
			ImageURL:    article.URLToImage,
			Category:    category,
			Status:      models.NewsStatusPublished,
			Published:   true,
			PublishDate: article.PublishedAt,
			ExternalID:  externalID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		news = append(news, newsArticle)
	}

	log.Info().Int("count", len(news)).Str("query", query).Msg("Successfully searched news articles")
	return news, nil
}

// determineCategory attempts to determine the category of a news article based on its content
func determineCategory(title, description string) models.NewsCategory {
	text := strings.ToLower(title + " " + description)

	categoryKeywords := map[models.NewsCategory][]string{
		models.NewsCategoryTechnology: {"tech", "technology", "software", "hardware", "app", "computer", "digital", "cyber", "ai", "artificial intelligence", "machine learning", "robot"},
		//models.NewsCategoryBusiness:      {"business", "company", "economy", "market", "stock", "finance", "investment", "startup", "entrepreneur"},
		models.NewsCategoryScience: {"science", "research", "study", "discovery", "space", "physics", "chemistry", "biology", "astronomy"},
		//models.NewsCategoryHealth:        {"health", "medical", "disease", "virus", "doctor", "hospital", "medicine", "covid", "vaccine", "treatment"},
		//models.NewsCategorySports:        {"sport", "game", "team", "player", "football", "soccer", "baseball", "basketball", "tennis", "golf", "olympics"},
		//models.NewsCategoryEntertainment: {"entertainment", "movie", "film", "music", "celebrity", "actor", "actress", "tv", "show", "concert", "festival", "award"},
	}

	// Count keyword matches for each category
	categoryScores := make(map[models.NewsCategory]int)

	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				categoryScores[category]++
			}
		}
	}

	// Find category with highest score
	var bestCategory models.NewsCategory = models.NewsCategoryTechnology
	highestScore := 0

	for category, score := range categoryScores {
		if score > highestScore {
			highestScore = score
			bestCategory = category
		}
	}

	return bestCategory
}
