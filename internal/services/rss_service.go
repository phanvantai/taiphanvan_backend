package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/mmcdole/gofeed"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
)

// RSSService handles fetching news from RSS feeds
type RSSService struct {
	cfg        config.RSSConfig
	httpClient *http.Client
	parser     *gofeed.Parser
}

// NewRSSService creates a new RSS feed service
func NewRSSService(cfg config.RSSConfig) (*RSSService, error) {
	if len(cfg.Feeds) == 0 {
		return nil, fmt.Errorf("no RSS feeds configured")
	}

	return &RSSService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		parser: gofeed.NewParser(),
	}, nil
}

// FetchNews fetches news articles from configured RSS feeds
func (s *RSSService) FetchNews(ctx context.Context, limit int) ([]models.News, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	maxLimit := 100
	if limit > maxLimit {
		limit = maxLimit // Cap at maximum
	}

	var allNews []models.News

	// Distribute the limit across feeds
	limitPerFeed := limit / len(s.cfg.Feeds)
	if limitPerFeed < 1 {
		limitPerFeed = 1
	}

	// Fetch from each feed
	for _, feed := range s.cfg.Feeds {
		feedNews, err := s.fetchFromFeed(ctx, feed, limitPerFeed)
		if err != nil {
			log.Error().Err(err).Str("feed_url", feed.URL).Msg("Failed to fetch news from RSS feed")
			continue // Continue with other feeds
		}

		allNews = append(allNews, feedNews...)
	}

	return allNews, nil
}

// fetchFromFeed fetches news articles from a single RSS feed
func (s *RSSService) fetchFromFeed(ctx context.Context, feed config.RSSFeed, limit int) ([]models.News, error) {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Execute request
	log.Info().Str("url", feed.URL).Str("name", feed.Name).Msg("Fetching news from RSS feed")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RSS feed returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse feed
	parsedFeed, err := s.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	// Process items into news articles
	var news []models.News
	newsCategory := mapCategoryString(feed.Category)

	// Cap the number of items to process
	itemCount := len(parsedFeed.Items)
	if itemCount > limit {
		itemCount = limit
	}

	for i := 0; i < itemCount; i++ {
		item := parsedFeed.Items[i]
		// Generate a unique slug
		titleSlug := slug.Make(item.Title)

		// Create an external ID for this item (format: rss-feedname-itemguid)
		feedName := slug.Make(feed.Name)
		itemID := item.GUID
		if itemID == "" {
			itemID = item.Link
		}

		// Create a compact itemID based on the URL path
		if strings.Contains(itemID, "http") {
			parsedURL, urlErr := url.Parse(itemID)
			if urlErr == nil {
				itemID = parsedURL.Path
			}
		}

		// Ensure the external ID doesn't exceed 100 characters
		externalID := fmt.Sprintf("rss-%s-%s", feedName, itemID)
		if len(externalID) > 100 {
			externalID = externalID[:100]
		}

		// Extract content from the feed item
		content := item.Content
		if content == "" {
			content = item.Description
		}

		// Create the news article
		publishDate := time.Now()
		if item.PublishedParsed != nil {
			publishDate = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishDate = *item.UpdatedParsed
		}

		// Get image URL if available
		imageURL := ""
		if item.Image != nil && item.Image.URL != "" {
			imageURL = item.Image.URL
		} else {
			// Try to find an enclosure that might be an image
			for _, enclosure := range item.Enclosures {
				if strings.HasPrefix(enclosure.Type, "image/") && enclosure.URL != "" {
					imageURL = enclosure.URL
					break
				}
			}
		}

		newsArticle := models.News{
			Title:       item.Title,
			Slug:        titleSlug,
			Content:     content,
			Summary:     item.Description,
			Source:      feed.Name,
			SourceURL:   item.Link,
			ImageURL:    imageURL,
			Category:    newsCategory,
			Status:      models.NewsStatusPublished,
			Published:   true,
			PublishDate: publishDate,
			ExternalID:  externalID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		news = append(news, newsArticle)
	}

	log.Info().Int("count", len(news)).Str("feed", feed.Name).Msg("Successfully fetched news articles from RSS")
	return news, nil
}

// mapCategoryString maps a string category to the NewsCategory type
func mapCategoryString(category string) models.NewsCategory {
	// Default to technology if no category is specified
	if category == "" {
		return models.NewsCategoryTechnology
	}

	// Convert to lowercase for case-insensitive comparison
	category = strings.ToLower(category)

	switch category {
	case "technology", "tech":
		return models.NewsCategoryTechnology
	case "science":
		return models.NewsCategoryScience
	default:
		return models.NewsCategory(category)
	}
}
