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
		Bool("auto_fetch_enabled", newsConfig.EnableAutoFetch).
		Dur("fetch_interval", newsConfig.FetchInterval).
		Msg("News fetcher configuration check")

	if !newsConfig.EnableAutoFetch {
		log.Info().Msg("Automatic news fetching is disabled")
		return
	}

	ticker := time.NewTicker(newsConfig.FetchInterval)

	go func() {
		log.Info().
			Dur("interval", newsConfig.FetchInterval).
			Msg("Starting automatic news fetcher background process")

		// Run immediately on startup
		fetchNewsFromAPI(newsConfig)

		// Then run on the scheduled interval
		for range ticker.C {
			fetchNewsFromAPI(newsConfig)
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

	// Begin transaction
	tx := database.DB.Begin()

	// Store each news article
	var savedCount int
	for _, article := range news {
		// Check if article already exists by external ID
		var existingCount int64
		if err := tx.Model(&models.News{}).Where("external_id = ?", article.ExternalID).Count(&existingCount).Error; err != nil {
			log.Error().Err(err).Str("external_id", article.ExternalID).Msg("Failed to check existing news")
			continue
		}

		if existingCount > 0 {
			continue // Skip existing articles
		}

		// Save article
		if err := tx.Create(&article).Error; err != nil {
			log.Error().Err(err).Str("title", article.Title).Msg("Failed to save news article")
			continue
		}

		savedCount++
	}

	// Commit transaction
	tx.Commit()

	log.Info().
		Int("total_fetched", len(news)).
		Int("saved", savedCount).
		Time("fetch_time", time.Now()).
		Msg("Successfully fetched and saved news articles")
}
