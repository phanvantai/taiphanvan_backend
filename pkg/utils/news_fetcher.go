package utils

import (
	"context"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/rs/zerolog/log"
)

// StartNewsFetcher starts the background process to automatically fetch news
func StartNewsFetcher(newsConfig services.NewsConfig) {
	log.Info().
		Bool("api_auto_fetch_enabled", newsConfig.EnableAutoFetch).
		Bool("rss_auto_fetch_enabled", newsConfig.RSSConfig.EnableAutoFetch).
		Dur("fetch_interval", newsConfig.FetchInterval).
		Msg("News fetcher configuration check")

	// Start API fetcher if enabled
	if newsConfig.EnableAutoFetch {
		startAPIFetcher(newsConfig)
	} else {
		log.Info().Msg("Automatic news API fetching is disabled")
	}

	// Start RSS fetcher if enabled
	if newsConfig.RSSConfig.EnableAutoFetch {
		startRSSFetcher(newsConfig)
	} else {
		log.Info().Msg("Automatic RSS fetching is disabled")
	}
}

// startAPIFetcher starts the background process to fetch news from the NewsAPI
func startAPIFetcher(newsConfig services.NewsConfig) {
	ticker := time.NewTicker(newsConfig.FetchInterval)

	go func() {
		log.Info().
			Dur("interval", newsConfig.FetchInterval).
			Msg("Starting automatic news API fetcher background process")

		// Run immediately on startup
		fetchNewsFromAPI(newsConfig)

		// Then run on the scheduled interval
		for range ticker.C {
			fetchNewsFromAPI(newsConfig)
		}
	}()
}

// startRSSFetcher starts the background process to fetch news from RSS feeds
func startRSSFetcher(newsConfig services.NewsConfig) {
	ticker := time.NewTicker(newsConfig.RSSConfig.FetchInterval)

	go func() {
		log.Info().
			Dur("interval", newsConfig.RSSConfig.FetchInterval).
			Msg("Starting automatic RSS fetcher background process")

		// Run immediately on startup
		fetchNewsFromRSS(newsConfig)

		// Then run on the scheduled interval
		for range ticker.C {
			fetchNewsFromRSS(newsConfig)
		}
	}()
}

// fetchNewsFromAPI fetches news articles from the external API and stores them in the database
func fetchNewsFromAPI(newsConfig services.NewsConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize news service
	newsService, err := services.NewNewsService(newsConfig.APIConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize news service for background fetching")
		return
	}

	// Define categories to fetch
	categories := []models.NewsCategory{
		models.NewsCategoryTechnology,
		models.NewsCategoryScience,
		// models.NewsCategoryBlockchain,
		// models.NewsCategoryDevelopment,
		// models.NewsCategoryWeb3,
	}

	// Fetch news from API
	log.Info().Msg("Fetching news from external API")
	news, err := newsService.FetchNews(ctx, categories, newsConfig.DefaultLimit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch news from external API")
		return
	}

	if len(news) == 0 {
		log.Info().Msg("No news articles found to import")
		return
	}

	// Store the fetched news articles
	saveNewsArticles(news)
}

// fetchNewsFromRSS fetches news articles from RSS feeds and stores them in the database
func fetchNewsFromRSS(newsConfig services.NewsConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize RSS service
	rssService, err := services.NewRSSService(newsConfig.RSSConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize RSS service for background fetching")
		return
	}

	// Fetch news from RSS feeds
	log.Info().Msg("Fetching news from RSS feeds")
	news, err := rssService.FetchNews(ctx, newsConfig.RSSConfig.DefaultLimit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch news from RSS feeds")
		return
	}

	if len(news) == 0 {
		log.Info().Msg("No news articles found in RSS feeds to import")
		return
	}

	// Store the fetched news articles
	saveNewsArticles(news)
}

// saveNewsArticles saves the news articles to the database
func saveNewsArticles(news []models.News) {
	// Don't use a single transaction for all articles to avoid
	// aborting the entire batch on a single error

	// Store each news article
	var savedCount int
	for _, article := range news {
		// Use a separate transaction for each article
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

		// If slug exists, skip this article
		if slugCount > 0 {
			log.Info().Str("slug", article.Slug).Msg("Skipping article with duplicate slug")
			tx.Rollback() // Clean rollback for skipped articles
			continue      // Skip articles with duplicate slugs
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

	log.Info().
		Int("total_fetched", len(news)).
		Int("saved", savedCount).
		Time("fetch_time", time.Now()).
		Msg("Successfully fetched and saved news articles")
}
