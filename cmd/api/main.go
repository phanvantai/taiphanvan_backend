package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/phanvantai/taiphanvan_backend/docs"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/handlers"
	"github.com/phanvantai/taiphanvan_backend/internal/logger"
	"github.com/phanvantai/taiphanvan_backend/internal/middleware"
	"github.com/phanvantai/taiphanvan_backend/internal/services"
	"github.com/phanvantai/taiphanvan_backend/pkg/utils"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           TaiPhanVan Blog API
// @version         1.0
// @description     A RESTful API for the TaiPhanVan personal blog platform with blog posts, user authentication, file management, and news features
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/phanvantai/taiphanvan_backend
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:9876
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	// Initialize barebones logger for startup errors
	initStartupLogger()

	// Load configuration (will automatically check for .env in root directory)
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger with proper configuration
	logger.Setup(cfg)

	// Set the Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize the database
	if err := database.Initialize(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close database")
		}
	}()

	// Set the JWT config for middleware
	middleware.SetConfig(cfg)

	// Start the token cleanup routine in the background
	utils.StartTokenCleanup()

	// Initialize Swagger documentation
	initSwagger()

	// Initialize the router
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add request ID middleware
	r.Use(requestIDMiddleware())

	// Add structured logger middleware
	r.Use(logger.GinMiddleware())

	// Configure CORS
	corsConfig := cors.DefaultConfig()

	// Add your custom domain to allowed origins
	customDomains := []string{"https://api.taiphanvan.dev", "https://taiphanvan.dev"}

	// Combine with existing allowed origins from config
	corsConfig.AllowOrigins = append(customDomains, cfg.CORS.AllowedOrigins...)

	// If in development mode, also allow localhost
	if cfg.Server.GinMode != "release" {
		corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, "http://localhost:*")
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	corsConfig.ExposeHeaders = []string{"Content-Length", "X-Request-ID"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

	r.Use(cors.New(corsConfig))

	// Initialize and apply rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute per IP
	rateLimiter.CleanupTask()                                  // Start the cleanup task

	// Start the news fetcher in background if enabled
	newsConfig := services.NewNewsConfig(cfg.NewsAPI, cfg.RSS)
	log.Info().
		Bool("api_auto_fetch_enabled", newsConfig.EnableAutoFetch).
		Bool("rss_auto_fetch_enabled", newsConfig.RSSConfig.EnableAutoFetch).
		Dur("api_fetch_interval", newsConfig.FetchInterval).
		Dur("rss_fetch_interval", newsConfig.RSSConfig.FetchInterval).
		Int("api_default_limit", newsConfig.DefaultLimit).
		Int("rss_default_limit", newsConfig.RSSConfig.DefaultLimit).
		Str("api_key_set", map[bool]string{true: "yes", false: "no"}[newsConfig.APIConfig.APIKey != ""]).
		Int("rss_feeds_configured", len(newsConfig.RSSConfig.Feeds)).
		Msg("News fetcher configuration")
	utils.StartNewsFetcher(newsConfig)

	// Define API routes with rate limiting
	setupRoutes(r, rateLimiter)

	// Create server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port),
		Handler: r,
	}

	// Serve in a goroutine so we can handle shutdown
	go func() {
		if cfg.TLS.Enabled {
			log.Info().
				Str("port", cfg.Server.Port).
				Str("mode", cfg.Server.GinMode).
				Msg("Server is running with TLS enabled")

			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start server")
			}
		} else {
			log.Info().
				Str("port", cfg.Server.Port).
				Str("mode", cfg.Server.GinMode).
				Msg("Server is running")

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start server")
			}
		}
	}()

	// Log Swagger URL with the correct protocol
	host := docs.SwaggerInfo.Host
	protocol := "http"
	if os.Getenv("RAILWAY_SERVICE_ID") != "" || os.Getenv("PRODUCTION") == "true" || cfg.TLS.Enabled {
		protocol = "https"
	}
	swaggerURL := fmt.Sprintf("%s://%s/swagger/index.html", protocol, host)
	log.Info().Str("url", swaggerURL).Msg("Swagger documentation available at")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Kill (no param) default sends syscall.SIGTERM
	// Kill -2 sends syscall.SIGINT
	// Kill -9 sends syscall.SIGKILL but can't be caught, so we don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Create a deadline to wait for in-flight requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}

// initStartupLogger initializes a basic logger for startup errors
func initStartupLogger() {
	// Set up a minimal console logger for startup
	consoleWriter := os.Stdout
	log.Logger = log.Output(consoleWriter)
}

// This function has been moved to the config package

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if a request ID was already set (e.g., by a load balancer or API gateway)
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID
			requestID = uuid.New().String()
		}

		// Set the request ID in the context and response headers
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// initSwagger initializes the Swagger documentation with the correct host
func initSwagger() {
	// Get host from environment or use default
	host := os.Getenv("API_HOST")
	if host == "" {
		// Check if we should use the custom domain
		customDomain := os.Getenv("CUSTOM_DOMAIN")
		if customDomain != "" {
			host = customDomain
		} else if railwayURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayURL != "" {
			// For Railway, use the PUBLIC_URL without scheme
			host = railwayURL
		} else if port := os.Getenv("API_PORT"); port != "" {
			host = fmt.Sprintf("localhost:%s", port)
		} else {
			host = "localhost:9876" // Default fallback
		}
	}

	// Set the Swagger info
	swaggerInfo := docs.SwaggerInfo
	swaggerInfo.Title = "TaiPhanVan Blog API"
	swaggerInfo.Description = "A RESTful API for the TaiPhanVan personal blog platform"
	swaggerInfo.Version = "1.0"
	swaggerInfo.Host = host
	swaggerInfo.BasePath = "/api"

	// Set the scheme based on environment
	isProduction := os.Getenv("RAILWAY_SERVICE_ID") != "" || os.Getenv("PRODUCTION") == "true"
	if isProduction || strings.HasPrefix(host, "api.taiphanvan.dev") {
		swaggerInfo.Schemes = []string{"https"}
	} else {
		swaggerInfo.Schemes = []string{"http"}
	}

	// Ensure the template variables are properly replaced in the Swagger JSON
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/api"

	log.Info().
		Str("title", swaggerInfo.Title).
		Str("version", swaggerInfo.Version).
		Str("host", host).
		Str("basePath", swaggerInfo.BasePath).
		Strs("schemes", swaggerInfo.Schemes).
		Msg("Swagger configuration initialized")
}

// setupRoutes configures all the routes for the API
func setupRoutes(r *gin.Engine, rateLimiter *middleware.RateLimiter) {
	// Define API routes
	api := r.Group("/api")
	{
		// Health check endpoint
		api.GET("/health", handlers.HealthCheck)

		// Apply rate limiting to all other API routes
		api.Use(rateLimiter.RateLimitMiddleware())

		// Public routes
		api.GET("/posts", handlers.GetPosts)
		api.GET("/posts/slug/:slug", handlers.GetPostBySlug)
		api.GET("/posts/:id/comments", handlers.GetCommentsByPostID)
		api.GET("/tags", handlers.GetAllTags)
		api.GET("/tags/popular", handlers.GetPopularTags)

		// News routes
		api.GET("/news", handlers.GetNews)
		api.GET("/news/slug/:slug", handlers.GetNewsBySlug)
		api.GET("/news/:id", handlers.GetNewsByID)
		api.GET("/news/:id/full-content", handlers.GetNewsFullContent)
		api.GET("/news/categories", handlers.GetNewsCategories)

		// Auth routes - stricter rate limiting for sensitive endpoints
		auth := api.Group("/auth")
		{
			// Create more restrictive rate limiter for auth endpoints
			authLimiter := middleware.NewRateLimiter(20, time.Minute) // 20 requests per minute per IP
			auth.Use(authLimiter.RateLimitMiddleware())

			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)
			auth.POST("/revoke", middleware.AuthMiddleware(), handlers.RevokeToken)
			auth.POST("/logout", middleware.AuthMiddleware(), handlers.Logout)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)
			protected.POST("/profile/avatar", handlers.UploadAvatar)

			// File routes for editor
			protected.POST("/files/upload", handlers.UploadFile)
			protected.POST("/files/delete", handlers.DeleteFile)

			// Post routes
			protected.POST("/posts", handlers.CreatePost)
			protected.PUT("/posts/:id", handlers.UpdatePost)
			protected.DELETE("/posts/:id", handlers.DeletePost)
			protected.GET("/posts/me", handlers.GetMyPosts) // New endpoint for dashboard
			protected.POST("/posts/:id/cover", handlers.UploadPostCover)
			protected.DELETE("/posts/:id/cover", handlers.DeletePostCover)
			protected.POST("/posts/:id/publish", handlers.PublishPost)
			protected.POST("/posts/:id/unpublish", handlers.UnpublishPost)
			protected.POST("/posts/:id/status", handlers.SetPostStatus)

			// Comment routes
			protected.POST("/posts/:id/comments", handlers.CreateComment)
			protected.PUT("/comments/:commentID", handlers.UpdateComment)
			protected.DELETE("/comments/:commentID", handlers.DeleteComment)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			// Admin-specific routes can be added here

			// News management routes
			admin.POST("/news", handlers.CreateNews)
			admin.PUT("/news/:id", handlers.UpdateNews)
			admin.DELETE("/news/:id", handlers.DeleteNews)
			admin.POST("/news/:id/status", handlers.SetNewsStatus)
			admin.POST("/news/fetch", handlers.FetchExternalNews)
			admin.POST("/news/fetch-rss", handlers.FetchRSSNews)
		}
	}

	// Add Swagger documentation endpoint with environment-aware configuration
	r.GET("/swagger/*any", func(c *gin.Context) {
		// Handle doc.json with the custom handler
		if c.Param("any") == "/doc.json" {
			handlers.SwaggerDocHandler(c)
			return
		}

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
				host = fmt.Sprintf("localhost:%s", port)
			} else {
				host = "localhost:9876" // Default fallback
			}
		}

		// Determine the correct protocol based on environment
		protocol := "http"
		if isProduction || strings.HasPrefix(host, "api.taiphanvan.dev") {
			protocol = "https"
		}

		// Configure Swagger with the correct URL
		swaggerURL := fmt.Sprintf("%s://%s/swagger/doc.json", protocol, host)
		log.Info().Str("swagger_url", swaggerURL).Msg("Configuring Swagger documentation URL")

		// Update Swagger info again to ensure it's properly set
		docs.SwaggerInfo.Host = host
		docs.SwaggerInfo.BasePath = "/api"

		ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL(swaggerURL),
			ginSwagger.DeepLinking(true),
			ginSwagger.DefaultModelsExpandDepth(1), // Show models with depth 1
			ginSwagger.DocExpansion("list"),        // Expand operation by default
		)(c)
	})
}
