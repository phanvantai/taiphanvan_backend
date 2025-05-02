package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Logging  LoggingConfig
	TLS      TLSConfig
	Admin    AdminConfig
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

// Load loads the configuration from environment variables or .env file
func Load(envFile string) (*Config, error) {
	// Try to load .env file if provided
	if envFile != "" {
		// First try direct path
		err := godotenv.Load(envFile)
		if err != nil {
			// If that fails, try to find it relative to working directory
			log.Warn().Err(err).Str("file", envFile).Msg("Failed to load .env file directly, trying relative path")

			// Get current working directory
			cwd, err := os.Getwd()
			if err == nil {
				relativePath := filepath.Join(cwd, envFile)
				err = godotenv.Load(relativePath)
				if err != nil {
					log.Warn().Err(err).Str("file", relativePath).Msg("Failed to load .env from relative path")
				}
			}

			// If we still fail, log but continue (we'll use env vars or defaults)
			if err != nil {
				log.Info().Msg("Continuing with environment variables and defaults")
			}
		}
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

	// Construct DSN
	dbConfig.DSN = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Name, dbConfig.Port, dbConfig.SSLMode)

	if dbConfig.Password != "" {
		dbConfig.DSN = fmt.Sprintf("%s password=%s", dbConfig.DSN, dbConfig.Password)
	}

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

	// Validate configuration - but provide better fallbacks for Docker environment
	if err := config.ValidateWithFallbacks(); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateWithFallbacks checks if all required configuration is present, with better Docker support
func (c *Config) ValidateWithFallbacks() error {
	// Check for Railway deployment
	if os.Getenv("RAILWAY") == "true" || os.Getenv("RAILWAY_SERVICE_ID") != "" {
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
			// This prevents invalid connection strings
			if c.Database.Host == "" {
				log.Info().Msg("Setting default database host for Railway")
				c.Database.Host = "localhost" // Default fallback
			}

			if c.Database.User == "" {
				log.Info().Msg("Setting default database user for Railway")
				c.Database.User = "postgres" // Common default for PostgreSQL
			}

			// Reconstruct the DSN with proper formatting
			c.Database.DSN = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
				c.Database.Host, c.Database.Port, c.Database.User, c.Database.Name, c.Database.SSLMode)

			if c.Database.Password != "" {
				c.Database.DSN = fmt.Sprintf("%s password=%s", c.Database.DSN, c.Database.Password)
			}

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

	// Database validation
	if c.Database.Host == "" {
		// For Docker Compose setups, try the service name as fallback
		if IsRunningInContainer() {
			log.Info().Msg("Running in container, using 'postgres' as default database host")
			c.Database.Host = "postgres"
		} else {
			return fmt.Errorf("DB_HOST is required")
		}
	}

	if c.Database.User == "" {
		// Common default for local development
		if IsRunningInContainer() {
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
		if os.Getenv("GIN_MODE") != "release" || IsRunningInContainer() {
			log.Warn().Msg("No JWT_SECRET provided. Using a default secret for development. DO NOT USE IN PRODUCTION!")
			c.JWT.Secret = "default_development_secret_replace_in_production"
		} else {
			return fmt.Errorf("JWT_SECRET is required in production mode")
		}
	} else if len(c.JWT.Secret) < 32 && os.Getenv("GIN_MODE") == "release" {
		return fmt.Errorf("JWT_SECRET should be at least 32 characters long in production mode")
	}

	// Reconstruct DSN with updated values
	c.Database.DSN = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
		c.Database.Host, c.Database.User, c.Database.Name, c.Database.Port, c.Database.SSLMode)

	if c.Database.Password != "" {
		c.Database.DSN = fmt.Sprintf("%s password=%s", c.Database.DSN, c.Database.Password)
	}

	return nil
}

// IsRunningInContainer checks if the application is running in a container
func IsRunningInContainer() bool {
	// This checks for the "/.dockerenv" file which is present in Docker containers
	_, err := os.Stat("/.dockerenv")
	return err == nil
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

// Default configuration values
func (c *Config) setDefaults() {
	// Server defaults
	if c.Server.Port == "" {
		c.Server.Port = getEnv("API_PORT", "9876")
	}
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
