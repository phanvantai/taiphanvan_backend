package database

import (
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// MigrateEnrichedContent creates the enriched_news_contents table
func MigrateEnrichedContent(db *gorm.DB) {
	log.Info().Msg("Running migration: Create enriched_news_contents table")
	err := db.AutoMigrate(&models.EnrichedNewsContent{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to migrate enriched_news_contents table")
	} else {
		log.Info().Msg("Successfully migrated enriched_news_contents table")
	}
}
