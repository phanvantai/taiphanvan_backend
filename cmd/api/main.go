package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/phanvantai/personal_blog_backend/docs"
	"github.com/phanvantai/personal_blog_backend/internal/config"
	"github.com/phanvantai/personal_blog_backend/internal/database"
	"github.com/phanvantai/personal_blog_backend/internal/handlers"
	"github.com/phanvantai/personal_blog_backend/internal/logger"
	"github.com/phanvantai/personal_blog_backend/internal/middleware"
	"github.com/phanvantai/personal_blog_backend/pkg/utils"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Personal Blog API
// @version         1.0
// @description     This is a REST API server for a personal blog.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/phanvantai/personal_blog_backend
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      ${API_HOST}
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	// Initialize barebones logger for startup errors
	initStartupLogger()

	// Check if we're running in a cloud environment
	var configPath string
	if isContainerized() || os.Getenv("RAILWAY_SERVICE_ID") != "" {
		log.Info().Msg("Running in a containerized/cloud environment")
		// In container or cloud, trust environment variables - no .env file needed
		configPath = ""
	} else {
		log.Info().Msg("Running in a local environment")
		// For local development, try to load from .env file
		configPath = "configs/.env"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Warn().Str("path", configPath).Msg("Config file not found, using environment variables")
			configPath = ""
		}
	}

	// Load configuration
	cfg, err := config.Load(configPath)
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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize and apply rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute per IP
	rateLimiter.CleanupTask()                                  // Start the cleanup task

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

// isContainerized checks if the application is running in a container
func isContainerized() bool {
	// Check for container environment indicators
	// 1. Check for /.dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 2. Check cgroup - common for Docker and other container engines
	if _, err := os.Stat("/proc/1/cgroup"); err == nil {
		data, err := os.ReadFile("/proc/1/cgroup")
		if err == nil && (
		// Look for container-related cgroup entries
		contains(string(data), "docker") ||
			contains(string(data), "kubepods") ||
			contains(string(data), "containerd")) {
			return true
		}
	}

	return false
}

// contains is a simple helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr) && s[0:len(substr)] == substr
}

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
		// For Railway, use the PUBLIC_URL without scheme
		if railwayURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayURL != "" {
			host = railwayURL
		} else if port := os.Getenv("API_PORT"); port != "" {
			host = fmt.Sprintf("localhost:%s", port)
		} else {
			host = "localhost:9876" // Default fallback
		}
	}

	// Set the host in the Swagger info
	swaggerInfo := docs.SwaggerInfo
	swaggerInfo.Host = host

	// Set the scheme based on environment
	isProduction := os.Getenv("RAILWAY_SERVICE_ID") != "" || os.Getenv("PRODUCTION") == "true"
	if isProduction {
		swaggerInfo.Schemes = []string{"https"}
	} else {
		swaggerInfo.Schemes = []string{"http"}
	}

	log.Info().
		Str("host", host).
		Strs("schemes", swaggerInfo.Schemes).
		Msg("Swagger configuration initialized")
}

// setupRoutes configures all the routes for the API
func setupRoutes(r *gin.Engine, rateLimiter *middleware.RateLimiter) {
	// Add health check endpoints
	r.GET("/health", handlers.HealthCheck)

	// Define API routes
	api := r.Group("/api")
	{
		// Apply rate limiting to all API routes
		api.Use(rateLimiter.RateLimitMiddleware())

		// Public routes
		api.GET("/posts", handlers.GetPosts)
		api.GET("/posts/slug/:slug", handlers.GetPostBySlug)
		api.GET("/posts/:id/comments", handlers.GetCommentsByPostID)
		api.GET("/tags", handlers.GetAllTags)
		api.GET("/tags/popular", handlers.GetPopularTags)

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

			// Post routes
			protected.POST("/posts", handlers.CreatePost)
			protected.PUT("/posts/:id", handlers.UpdatePost)
			protected.DELETE("/posts/:id", handlers.DeletePost)

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
		}
	}

	// Add Swagger documentation endpoint with environment-aware configuration
	r.GET("/swagger/*any", func(c *gin.Context) {
		// Determine if we're running in Railway or other production environment
		isProduction := os.Getenv("RAILWAY_SERVICE_ID") != "" || os.Getenv("PRODUCTION") == "true"

		// Get host from environment or use default
		host := os.Getenv("API_HOST")
		if host == "" {
			// For Railway, use the PUBLIC_URL without scheme
			if railwayURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayURL != "" {
				host = railwayURL
			} else if port := os.Getenv("API_PORT"); port != "" {
				host = fmt.Sprintf("localhost:%s", port)
			} else {
				host = "localhost:9876" // Default fallback
			}
		}

		// Determine the correct protocol based on environment
		protocol := "http"
		if isProduction {
			protocol = "https"
		}

		// Configure Swagger with the correct URL
		swaggerURL := fmt.Sprintf("%s://%s/swagger/doc.json", protocol, host)
		log.Info().Str("swagger_url", swaggerURL).Msg("Configuring Swagger documentation URL")

		ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL(swaggerURL),
			ginSwagger.DeepLinking(true),
		)(c)
	})
}
