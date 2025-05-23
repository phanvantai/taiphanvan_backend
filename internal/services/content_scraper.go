package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
)

// ContentScraper handles fetching full content from news source URLs
type ContentScraper struct {
	httpClient *http.Client
}

// NewContentScraper creates a new content scraper service
func NewContentScraper() *ContentScraper {
	return &ContentScraper{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// IsTruncated checks if content appears to be truncated
func IsTruncated(content string) bool {
	// Check for common truncation patterns
	truncationPatterns := []string{
		"[+",        // NewsAPI style: "[+1234 chars]"
		"...",       // Common ellipsis
		"…",         // Unicode ellipsis
		"Read more", // Common text
		"&#8230;",   // HTML entity for ellipsis
		"[&#8230;]", // HTML entity for ellipsis in brackets
	}

	for _, pattern := range truncationPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// ExtractTruncationInfo extracts information about how much content is truncated
// Returns (isTruncated bool, truncatedChars int, truncationPattern string)
func ExtractTruncationInfo(content string) (bool, int, string) {
	// Check for NewsAPI style truncation [+1234 chars]
	re := regexp.MustCompile(`\[\+(\d+) chars\]`)
	match := re.FindStringSubmatch(content)

	if len(match) > 1 {
		// Found a match like [+1234 chars]
		charCount := 0
		fmt.Sscanf(match[1], "%d", &charCount)
		return true, charCount, match[0]
	}

	// Check for HTML entity ellipsis
	if strings.Contains(content, "&#8230;") {
		if strings.Contains(content, "[&#8230;]") {
			return true, 0, "[&#8230;]"
		}
		return true, 0, "&#8230;"
	}

	// Check for other truncation patterns
	if strings.Contains(content, "...") || strings.Contains(content, "…") {
		return true, 0, "..."
	}

	return false, 0, ""
}

// FetchFullContent attempts to fetch the full content of a news article
// Note: This is a simple implementation and may not work for all sources
// due to different website structures, paywalls, etc.
func (s *ContentScraper) FetchFullContent(ctx context.Context, sourceURL string) (string, error) {
	if sourceURL == "" {
		return "", errors.New("source URL is empty")
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Execute request
	log.Info().Str("url", sourceURL).Msg("Fetching full content from source")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("source returned non-OK status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Simple extraction of content
	// Note: This is a very simplified approach and will not work well
	// for most modern websites. A production solution would use a more
	// sophisticated approach like Readability algorithms, proper HTML parsing, etc.
	content := string(body)

	// Simple cleanup
	content = cleanupContent(content)

	return content, nil
}

// EnrichNewsContent enhances a news article by adding full content information
func (s *ContentScraper) EnrichNewsContent(ctx context.Context, news *models.News) (*models.EnrichedNewsContent, error) {
	if news == nil {
		return nil, errors.New("news article is nil")
	}

	// Check if content appears to be truncated
	isTruncated, chars, pattern := ExtractTruncationInfo(news.Content)

	enriched := &models.EnrichedNewsContent{
		NewsID:             news.ID,
		OriginalContent:    news.Content,
		IsTruncated:        isTruncated,
		TruncatedChars:     chars,
		TruncationPattern:  pattern,
		SourceURL:          news.SourceURL,
		TruncationDetected: time.Now(),
	}

	// Don't attempt to fetch content if not truncated
	if !isTruncated {
		return enriched, nil
	}

	// Attempt to fetch full content
	fullContent, err := s.FetchFullContent(ctx, news.SourceURL)
	if err != nil {
		log.Error().Err(err).Uint("newsID", news.ID).Str("sourceURL", news.SourceURL).Msg("Failed to fetch full content")
		enriched.FetchError = err.Error()
		return enriched, nil // Return what we have even if fetch failed
	}

	enriched.FullContent = fullContent
	enriched.LastFetched = time.Now()

	return enriched, nil
}

// cleanupContent performs basic cleanup on scraped content
func cleanupContent(content string) string {
	// This is a very simplified cleanup approach
	// In a production system, you would use proper HTML parsing and extraction

	// Remove script and style tags and their content
	scriptPattern := regexp.MustCompile(`<script[\s\S]*?</script>`)
	content = scriptPattern.ReplaceAllString(content, "")

	stylePattern := regexp.MustCompile(`<style[\s\S]*?</style>`)
	content = stylePattern.ReplaceAllString(content, "")

	// Remove HTML tags
	htmlTagPattern := regexp.MustCompile(`<[^>]*>`)
	content = htmlTagPattern.ReplaceAllString(content, " ")

	// Remove excess whitespace
	whitespacePattern := regexp.MustCompile(`\s+`)
	content = whitespacePattern.ReplaceAllString(content, " ")

	return strings.TrimSpace(content)
}
