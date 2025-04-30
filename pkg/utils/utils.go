package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/phanvantai/personal_blog_backend/internal/database"
	"github.com/phanvantai/personal_blog_backend/internal/models"
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

// StartTokenCleanup starts a goroutine to periodically clean up expired blacklisted tokens
func StartTokenCleanup() {
	go func() {
		ticker := time.NewTicker(12 * time.Hour) // Run cleanup every 12 hours
		defer ticker.Stop()

		// Run an immediate cleanup on startup
		cleanupTokens()

		// Then run on the ticker schedule
		for range ticker.C {
			cleanupTokens()
		}
	}()
	log.Println("Token cleanup routine started")
}

// cleanupTokens removes expired tokens from the blacklist
func cleanupTokens() {
	result := database.DB.Where("expires_at < ?", time.Now()).Delete(&models.BlacklistedToken{})
	if result.Error != nil {
		log.Printf("Error cleaning up expired tokens: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired tokens", result.RowsAffected)
	}
}
