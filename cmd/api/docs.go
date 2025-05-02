// swagger documentation for the Personal Blog Backend API
package main

import (
	_ "github.com/phanvantai/personal_blog_backend/docs" // Import the docs package
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

// @host      localhost:9876
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

// @tag.name Auth
// @tag.description Authentication operations

// @tag.name Posts
// @tag.description Blog post operations

// @tag.name Comments
// @tag.description Comment operations

// @tag.name Tags
// @tag.description Tag operations

// @tag.name Users
// @tag.description User operations
