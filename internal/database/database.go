package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the database connection with proper connection pooling
func Initialize(cfg *config.Config) error {
	logLevel := logger.Info
	if cfg.Server.GinMode == "release" {
		logLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	// Use the DSN from config, which is already handled in config.go for Railway
	dsn := cfg.Database.DSN

	// Log connection attempt (without exposing credentials)
	hostInfo := "using DATABASE_URL"
	if os.Getenv("DATABASE_URL") == "" {
		hostInfo = fmt.Sprintf("using DSN with host=%s, dbname=%s", cfg.Database.Host, cfg.Database.Name)
	}
	log.Printf("Attempting to connect to database (%s)", hostInfo)

	// Make sure we have a reasonable connection timeout
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to open the database connection with context
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disables implicit prepared statement usage
	}), gormConfig)

	if err != nil {
		// Enhanced error reporting for easier debugging
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // Maximum idle time for a connection

	log.Println("Database connected successfully")

	// Auto migrate database schemas
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	// Create default admin user if enabled
	if cfg.Admin.CreateDefaultAdmin {
		if err := CreateDefaultAdminUser(cfg); err != nil {
			return fmt.Errorf("failed to create default admin user: %w", err)
		}
	}

	// Create default editor user if enabled
	if cfg.Editor.CreateDefaultEditor {
		if err := CreateDefaultEditorUser(cfg); err != nil {
			return fmt.Errorf("failed to create default editor user: %w", err)
		}
	}

	return nil
}

// autoMigrate automatically migrates the database schema
func autoMigrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Tag{},
		&models.Comment{},
		&models.BlacklistedToken{},
		&models.RefreshToken{},        // Add RefreshToken model
		&models.News{},                // Add News model
		&models.EnrichedNewsContent{}, // Add EnrichedNewsContent model
	)
	if err != nil {
		return err
	}

	log.Println("Database migration completed")
	return nil
}

// CreateDefaultAdminUser creates a default admin user if no admin exists
func CreateDefaultAdminUser(cfg *config.Config) error {
	// Check if admin user already exists
	var count int64
	if err := DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check for existing admin: %w", err)
	}

	// If admin exists, return early
	if count > 0 {
		log.Println("Admin user already exists, skipping default admin creation")
		return nil
	}

	// Get admin credentials from config or environment variables
	adminUsername := cfg.Admin.Username
	adminEmail := cfg.Admin.Email
	adminPassword := cfg.Admin.Password

	// Fall back to environment variables if not in config
	if adminUsername == "" {
		adminUsername = os.Getenv("DEFAULT_ADMIN_USERNAME")
		if adminUsername == "" {
			adminUsername = "admin" // Fallback default
		}
	}
	if adminEmail == "" {
		adminEmail = os.Getenv("DEFAULT_ADMIN_EMAIL")
		if adminEmail == "" {
			adminEmail = "admin@admin.com" // Fallback default
		}
	}
	if adminPassword == "" {
		adminPassword = os.Getenv("DEFAULT_ADMIN_PASSWORD")
		if adminPassword == "" {
			adminPassword = "securePassword123" // Fallback default
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user
	adminUser := models.User{
		Username:  adminUsername,
		Email:     adminEmail,
		Password:  string(hashedPassword),
		FirstName: "System",
		LastName:  "Admin",
		Role:      "admin",
	}

	if result := DB.Create(&adminUser); result.Error != nil {
		return fmt.Errorf("failed to create admin user: %w", result.Error)
	}

	log.Printf("Default admin user created with username: %s", adminUsername)
	return nil
}

// CreateDefaultEditorUser creates a default editor user if no editor exists
func CreateDefaultEditorUser(cfg *config.Config) error {
	// Check if editor user already exists
	var count int64
	if err := DB.Model(&models.User{}).Where("role = ?", "editor").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check for existing editor: %w", err)
	}

	// If editor exists, return early
	if count > 0 {
		log.Println("Editor user already exists, skipping default editor creation")
		return nil
	}

	// Get editor credentials from config or environment variables
	editorUsername := cfg.Editor.Username
	editorEmail := cfg.Editor.Email
	editorPassword := cfg.Editor.Password

	// Fall back to environment variables if not in config
	if editorUsername == "" {
		editorUsername = os.Getenv("DEFAULT_EDITOR_USERNAME")
		if editorUsername == "" {
			editorUsername = "editor" // Fallback default
		}
	}
	if editorEmail == "" {
		editorEmail = os.Getenv("DEFAULT_EDITOR_EMAIL")
		if editorEmail == "" {
			editorEmail = "editor@editor.com" // Fallback default
		}
	}
	if editorPassword == "" {
		editorPassword = os.Getenv("DEFAULT_EDITOR_PASSWORD")
		if editorPassword == "" {
			editorPassword = "securePassword123" // Fallback default
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(editorPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash editor password: %w", err)
	}

	// Create editor user
	editorUser := models.User{
		Username:  editorUsername,
		Email:     editorEmail,
		Password:  string(hashedPassword),
		FirstName: "Content",
		LastName:  "Editor",
		Role:      "editor",
	}

	if result := DB.Create(&editorUser); result.Error != nil {
		return fmt.Errorf("failed to create editor user: %w", result.Error)
	}

	log.Printf("Default editor user created with username: %s", editorUsername)
	return nil
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}
