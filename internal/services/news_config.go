package services

import (
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/config"
)

// NewsConfig holds the configuration for the news fetcher service
type NewsConfig struct {
	APIConfig       config.NewsAPIConfig
	DefaultLimit    int
	FetchInterval   time.Duration
	EnableAutoFetch bool
}

// NewNewsConfig creates a new NewsConfig from the application config
func NewNewsConfig(cfg config.NewsAPIConfig) NewsConfig {
	return NewsConfig{
		APIConfig:       cfg,
		DefaultLimit:    cfg.DefaultLimit,
		FetchInterval:   cfg.FetchInterval,
		EnableAutoFetch: cfg.EnableAutoFetch,
	}
}
