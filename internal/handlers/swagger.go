package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/docs"
	"github.com/rs/zerolog/log"
)

// SwaggerDocHandler handles the Swagger doc.json endpoint to ensure template variables are replaced
func SwaggerDocHandler(c *gin.Context) {
	// Determine if we're running in Railway or other production environment
	isProduction := os.Getenv("RAILWAY_SERVICE_ID") != "" || os.Getenv("PRODUCTION") == "true"

	// Get host from environment or use default
	host := os.Getenv("API_HOST")
	if host == "" {
		// Check if we should use the custom domain
		customDomain := os.Getenv("CUSTOM_DOMAIN")
		if customDomain != "" {
			host = customDomain
		} else if railwayURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayURL != "" {
			host = railwayURL
		} else if port := os.Getenv("API_PORT"); port != "" {
			host = "localhost:" + port
		} else {
			host = "localhost:9876" // Default fallback
		}
	}

	// Set the Swagger info
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/api"

	// Set the scheme based on environment
	if isProduction || strings.HasPrefix(host, "api.taiphanvan.dev") {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// Get the Swagger JSON
	doc := docs.SwaggerInfo.ReadDoc()

	// Replace any remaining template variables
	doc = strings.ReplaceAll(doc, "\"host\": \"{{.Host}}\"", fmt.Sprintf("\"host\": \"%s\"", host))
	doc = strings.ReplaceAll(doc, "\"basePath\": \"{{.BasePath}}\"", "\"basePath\": \"/api\"")

	// Also replace any other occurrences of template variables in the document
	doc = strings.ReplaceAll(doc, "http://{{.Host}}{{.BasePath}}", fmt.Sprintf("http://%s/api", host))
	doc = strings.ReplaceAll(doc, "https://{{.Host}}{{.BasePath}}", fmt.Sprintf("https://%s/api", host))

	log.Info().
		Str("host", host).
		Str("basePath", "/api").
		Msg("Serving Swagger documentation with replaced template variables")

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, doc)
}
