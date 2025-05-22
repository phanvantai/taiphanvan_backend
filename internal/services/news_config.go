package services

import (
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/config"
)

// NewsConfig holds the configuration for the news fetcher service
type NewsConfig struct {
	APIConfig       config.NewsAPIConfig
	RSSConfig       config.RSSConfig
	DefaultLimit    int
	FetchInterval   time.Duration
	EnableAutoFetch bool
}

// NewNewsConfig creates a new NewsConfig from the application config
func NewNewsConfig(apiCfg config.NewsAPIConfig, rssCfg config.RSSConfig) NewsConfig {
	return NewsConfig{
		APIConfig:       apiCfg,
		RSSConfig:       rssCfg,
		DefaultLimit:    apiCfg.DefaultLimit,
		FetchInterval:   apiCfg.FetchInterval,
		EnableAutoFetch: apiCfg.EnableAutoFetch,
	}
}
