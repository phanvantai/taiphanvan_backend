// Package utils provides utility functions used across the application
package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/database"
)

// FormatDate formats a time.Time to a human-readable date string
func FormatDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

// TruncateText truncates text to a specified length and adds ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")

	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// PrettyJSON converts an interface to a formatted JSON string
func PrettyJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(jsonBytes)
}

// ExtractExcerpt generates an excerpt from content if none is provided
func ExtractExcerpt(content string, maxLength int) string {
	// Remove HTML tags (very basic implementation)
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")

	return TruncateText(content, maxLength)
}

// StartTokenCleanup starts a background routine to clean up expired tokens
func StartTokenCleanup() {
	go func() {
		for {
			// Run cleanup every hour
			time.Sleep(1 * time.Hour)
			CleanupExpiredTokens()
		}
	}()
	log.Println("Token cleanup routine started")
}

// CleanupExpiredTokens removes expired tokens from the database
func CleanupExpiredTokens() {
	// Ensure database is initialized
	if database.DB == nil {
		log.Println("Database not initialized, skipping token cleanup")
		return
	}

	now := time.Now()

	// Clean up blacklisted tokens
	if result := database.DB.Exec("DELETE FROM blacklisted_tokens WHERE expires_at < ?", now); result.Error != nil {
		log.Printf("Error cleaning up blacklisted tokens: %v", result.Error)
	} else {
		if result.RowsAffected > 0 {
			log.Printf("Cleaned up %d expired blacklisted tokens", result.RowsAffected)
		}
	}

	// Clean up refresh tokens
	if result := database.DB.Exec("DELETE FROM refresh_tokens WHERE expires_at < ? OR revoked = true", now); result.Error != nil {
		log.Printf("Error cleaning up refresh tokens: %v", result.Error)
	} else {
		if result.RowsAffected > 0 {
			log.Printf("Cleaned up %d expired refresh tokens", result.RowsAffected)
		}
	}
}
