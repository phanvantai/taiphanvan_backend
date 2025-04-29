package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
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
