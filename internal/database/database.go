package database

import (
	"fmt"
	"log"
	"os"

	"github.com/phanvantai/personal_blog_backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize sets up the database connection
func Initialize() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Make sure we have the required database name - use blog_db as fallback
	if dbName == "" {
		dbName = "blog_db"
		log.Println("DB_NAME not set in environment, using default: blog_db")
	}

	// Construct the connection string
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbName, dbPort)

	// Only add password parameter if one is set
	if dbPass != "" {
		dsn = fmt.Sprintf("%s password=%s", dsn, dbPass)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
	autoMigrate()
}

// autoMigrate automatically migrates the database schema
func autoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Tag{},
		&models.Comment{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")
}
