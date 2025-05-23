package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	CORS       CORSConfig
	Logging    LoggingConfig
	TLS        TLSConfig
	Admin      AdminConfig
	Editor     EditorConfig
	Cloudinary CloudinaryConfig
	NewsAPI    NewsAPIConfig
	RSS        RSSConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	DSN      string // Connection string, computed from other fields
}

// JWTConfig holds all JWT-related configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile string
	KeyFile  string
	Enabled  bool
}

// AdminConfig holds configuration for the default admin user
type AdminConfig struct {
	CreateDefaultAdmin bool
	Username           string
	Email              string
	Password           string
}

// EditorConfig holds configuration for the default editor user
type EditorConfig struct {
	CreateDefaultEditor bool
	Username            string
	Email               string
	Password            string
}

// CloudinaryConfig holds configuration for Cloudinary
type CloudinaryConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string
}

// NewsAPIConfig holds configuration for NewsAPI
type NewsAPIConfig struct {
	BaseURL         string
	APIKey          string
	DefaultLimit    int
	FetchInterval   time.Duration
	EnableAutoFetch bool
}

// RSSFeed holds configuration for a single RSS feed
type RSSFeed struct {
	Name     string
	URL      string
	Category string
}

// RSSConfig holds configuration for RSS feeds
type RSSConfig struct {
	Feeds           []RSSFeed
	DefaultLimit    int
	FetchInterval   time.Duration
	EnableAutoFetch bool
}

// Load loads the configuration from environment variables or .env file
func Load(_ string) (*Config, error) {
	// Skip .env file loading in containerized environments (including Railway)
	// as they should use environment variables directly
	if !IsRunningInContainer() && !IsRunningOnRailway() {
		// Only try to load from the root .env file
		envFile := ".env"
		if err := godotenv.Load(envFile); err == nil {
			log.Info().Str("file", envFile).Msg("Loaded environment from .env file")
		} else {
			log.Info().Msg("No .env file found. Continuing with environment variables and defaults")
		}
	} else {
		log.Info().
			Bool("container", IsRunningInContainer()).
			Bool("railway", IsRunningOnRailway()).
			Msg("Running in containerized/cloud environment, using environment variables directly")
	}

	config := &Config{}

	// Load server config
	config.Server = ServerConfig{
		Port:    getEnv("API_PORT", "9876"),
		GinMode: getEnv("GIN_MODE", "debug"),
	}

	// Load database config
	dbConfig := DatabaseConfig{
		Host:     getEnv("DB_HOST", ""),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASS", ""),
		Name:     getEnv("DB_NAME", "blog_db"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Initial DSN construction (may be overridden in ValidateWithFallbacks)
	dbConfig.DSN = constructDSN(
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	config.Database = dbConfig

	// Load JWT config
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	config.JWT = JWTConfig{
		Secret:        getEnv("JWT_SECRET", ""),
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}

	// Load CORS config
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	origins := strings.Split(corsOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	config.CORS = CORSConfig{
		AllowedOrigins: origins,
	}

	// Load logging config
	config.Logging = LoggingConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "text"),
	}

	// Load TLS config
	certFile := getEnv("TLS_CERT_FILE", "")
	keyFile := getEnv("TLS_KEY_FILE", "")
	config.TLS = TLSConfig{
		CertFile: certFile,
		KeyFile:  keyFile,
		Enabled:  certFile != "" && keyFile != "",
	}

	// Load admin config
	config.Admin = AdminConfig{
		CreateDefaultAdmin: GetEnvBool("CREATE_DEFAULT_ADMIN", false),
		Username:           getEnv("DEFAULT_ADMIN_USERNAME", ""),
		Email:              getEnv("DEFAULT_ADMIN_EMAIL", ""),
		Password:           getEnv("DEFAULT_ADMIN_PASSWORD", ""),
	}

	// Load editor config
	config.Editor = EditorConfig{
		CreateDefaultEditor: GetEnvBool("CREATE_DEFAULT_EDITOR", false),
		Username:            getEnv("DEFAULT_EDITOR_USERNAME", ""),
		Email:               getEnv("DEFAULT_EDITOR_EMAIL", ""),
		Password:            getEnv("DEFAULT_EDITOR_PASSWORD", ""),
	}

	// Load Cloudinary config
	config.Cloudinary = CloudinaryConfig{
		CloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
		APIKey:       getEnv("CLOUDINARY_API_KEY", ""),
		APISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
		UploadFolder: getEnv("CLOUDINARY_UPLOAD_FOLDER", "blog_images"),
	}

	// Load NewsAPI config
	fetchInterval, err := time.ParseDuration(getEnv("NEWS_API_FETCH_INTERVAL", "1h"))
	if err != nil {
		fetchInterval = 1 * time.Hour // Default to 1 hour if invalid
	}

	defaultLimit, err := strconv.Atoi(getEnv("NEWS_API_DEFAULT_LIMIT", "10"))
	if err != nil {
		defaultLimit = 10 // Default to 10 if invalid
	}

	config.NewsAPI = NewsAPIConfig{
		BaseURL:         getEnv("NEWS_API_BASE_URL", "https://newsapi.org/v2"),
		APIKey:          getEnv("NEWS_API_KEY", ""),
		DefaultLimit:    defaultLimit,
		FetchInterval:   fetchInterval,
		EnableAutoFetch: GetEnvBool("NEWS_API_ENABLE_AUTO_FETCH", false),
	}

	// Load RSS config
	rssFetchInterval, err := time.ParseDuration(getEnv("RSS_FETCH_INTERVAL", "1h"))
	if err != nil {
		rssFetchInterval = 1 * time.Hour // Default to 1 hour if invalid
	}

	rssDefaultLimit, err := strconv.Atoi(getEnv("RSS_DEFAULT_LIMIT", "10"))
	if err != nil {
		rssDefaultLimit = 10 // Default to 10 if invalid
	}

	// Parse RSS feeds from environment variable
	// Format: NAME1=URL1=CATEGORY1,NAME2=URL2=CATEGORY2,...
	var rssFeeds []RSSFeed
	rssEnv := getEnv("RSS_FEEDS", "")
	if rssEnv != "" {
		feedsStr := strings.Split(rssEnv, ",")
		for _, feedStr := range feedsStr {
			parts := strings.Split(feedStr, "=")
			if len(parts) >= 2 {
				feed := RSSFeed{
					Name: parts[0],
					URL:  parts[1],
				}
				if len(parts) >= 3 {
					feed.Category = parts[2]
				} else {
					feed.Category = "technology" // Default category
				}
				rssFeeds = append(rssFeeds, feed)
			}
		}
	}

	config.RSS = RSSConfig{
		Feeds:           rssFeeds,
		DefaultLimit:    rssDefaultLimit,
		FetchInterval:   rssFetchInterval,
		EnableAutoFetch: GetEnvBool("RSS_ENABLE_AUTO_FETCH", false),
	}

	// Validate configuration and apply environment-specific fallbacks
	if err := config.ValidateWithFallbacks(); err != nil {
		return nil, err
	}

	return config, nil
}

// constructDSN creates a PostgreSQL connection string from individual parameters
func constructDSN(host, port, user, password, dbname, sslmode string) string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		host, port, user, dbname, sslmode)

	if password != "" {
		dsn = fmt.Sprintf("%s password=%s", dsn, password)
	}

	return dsn
}

// ValidateWithFallbacks checks if all required configuration is present, with better Docker support
func (c *Config) ValidateWithFallbacks() error {
	// Check for Railway deployment
	if IsRunningOnRailway() {
		log.Info().Msg("Running on Railway.app, applying platform-specific configuration")

		// Use Railway's PORT environment variable for the server
		if port := os.Getenv("PORT"); port != "" {
			log.Info().Str("port", port).Msg("Using PORT from Railway environment")
			c.Server.Port = port
		}

		// Use DATABASE_URL if available (provided by Railway)
		if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
			log.Info().Msg("Using DATABASE_URL from Railway")
			c.Database.DSN = dbURL
		} else {
			log.Warn().Msg("DATABASE_URL not found in Railway environment, checking for individual database variables")

			// For Railway, we'll set default database connection values if they're not provided
			if c.Database.Host == "" {
				log.Info().Msg("Setting default database host for Railway")
				c.Database.Host = "localhost" // Default fallback
			}

			if c.Database.User == "" {
				log.Info().Msg("Setting default database user for Railway")
				c.Database.User = "postgres" // Common default for PostgreSQL
			}

			// Reconstruct the DSN with proper formatting
			c.Database.DSN = constructDSN(
				c.Database.Host,
				c.Database.Port,
				c.Database.User,
				c.Database.Password,
				c.Database.Name,
				c.Database.SSLMode,
			)

			log.Info().Str("host", c.Database.Host).Str("dbname", c.Database.Name).
				Msg("Using constructed database connection string")
		}

		// For Railway, ensure we have a JWT secret
		if c.JWT.Secret == "" {
			log.Warn().Msg("No JWT_SECRET provided on Railway. Using a generated secret. It's recommended to set a persistent JWT_SECRET.")
			c.JWT.Secret = generateTemporarySecret()
		}

		return nil
	}

	// Check for Render.com deployment
	if os.Getenv("RENDER") == "true" {
		log.Info().Msg("Running on Render.com, applying platform-specific configuration")
		return nil // Render provides all necessary environment variables
	}

	// Handle Docker/container environment (local or other deployment)
	isContainer := IsRunningInContainer()

	// Database validation
	if c.Database.Host == "" {
		if isContainer {
			log.Info().Msg("Running in container, using 'postgres' as default database host")
			c.Database.Host = "postgres"
		} else {
			return fmt.Errorf("DB_HOST is required")
		}
	}

	if c.Database.User == "" {
		if isContainer {
			log.Info().Msg("Running in container, using 'bloguser' as default database user")
			c.Database.User = "bloguser"
			c.Database.Password = "blogpassword" // Default password from docker-compose
		} else {
			return fmt.Errorf("DB_USER is required")
		}
	}

	// JWT validation with better fallback for development
	if c.JWT.Secret == "" {
		// In development or container, we can use a default for convenience
		if os.Getenv("GIN_MODE") != "release" || isContainer {
			log.Warn().Msg("No JWT_SECRET provided. Using a default secret for development. DO NOT USE IN PRODUCTION!")
			c.JWT.Secret = "default_development_secret_replace_in_production"
		} else {
			return fmt.Errorf("JWT_SECRET is required in production mode")
		}
	} else if len(c.JWT.Secret) < 32 && os.Getenv("GIN_MODE") == "release" {
		return fmt.Errorf("JWT_SECRET should be at least 32 characters long in production mode")
	}

	// Reconstruct DSN with updated values
	c.Database.DSN = constructDSN(
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)

	return nil
}

// IsRunningInContainer checks if the application is running in a container
func IsRunningInContainer() bool {
	// 1. Check for /.dockerenv file (Docker)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 2. Check cgroup - common for Docker and other container engines
	if _, err := os.Stat("/proc/1/cgroup"); err == nil {
		data, err := os.ReadFile("/proc/1/cgroup")
		if err == nil {
			content := string(data)
			if strings.Contains(content, "docker") ||
				strings.Contains(content, "kubepods") ||
				strings.Contains(content, "containerd") {
				return true
			}
		}
	}

	return false
}

// IsRunningOnRailway checks if the application is running on Railway.app
func IsRunningOnRailway() bool {
	return os.Getenv("RAILWAY") == "true" || os.Getenv("RAILWAY_SERVICE_ID") != ""
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvBool gets a boolean environment variable
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return b
}

// generateTemporarySecret creates a more secure temporary secret
func generateTemporarySecret() string {
	// This is still not ideal for production, but better than a hardcoded value
	// It creates a unique secret per application start
	timestamp := time.Now().UnixNano()
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	return fmt.Sprintf("temporary_secret_%s_%d_%d", hostname, pid, timestamp)
}
