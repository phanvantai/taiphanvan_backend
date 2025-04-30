package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/phanvantai/personal_blog_backend/internal/database"
	"github.com/phanvantai/personal_blog_backend/internal/handlers"
	"github.com/phanvantai/personal_blog_backend/internal/middleware"
	"github.com/phanvantai/personal_blog_backend/pkg/utils"
)

func main() {
	// Load the .env file
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize the database
	database.Initialize()

	// Start the token cleanup routine
	utils.StartTokenCleanup()

	// Set up Gin
	r := gin.Default()

	// Handle CORS using Gin's native CORS middleware
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*" // Default to allow all origins if not specified
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Define API routes
	api := r.Group("/api")
	{
		// Public routes
		api.GET("/posts", handlers.GetPosts)
		api.GET("/posts/slug/:slug", handlers.GetPostBySlug)         // Changed path to avoid conflict
		api.GET("/posts/:id/comments", handlers.GetCommentsByPostID) // Changed parameter name to 'id'
		api.GET("/tags", handlers.GetAllTags)
		api.GET("/tags/popular", handlers.GetPopularTags)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", middleware.AuthMiddleware(), handlers.Logout) // Protected logout endpoint
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
			protected.POST("/posts/:id/comments", handlers.CreateComment) // Changed parameter name to 'id'
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

	// Get port from environment or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
