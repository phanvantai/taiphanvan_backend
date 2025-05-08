// Package models provides data models for the API
package models

import "time"

// SwaggerDeletedAt is a custom type for Swagger documentation
// @Description A timestamp for soft-deleted records (null if not deleted)
type SwaggerDeletedAt struct {
	Time  time.Time `json:"time,omitempty"`
	Valid bool      `json:"valid,omitempty"`
}

// SwaggerStandardResponse represents a standard API response
// @Description A standard API response format
type SwaggerStandardResponse struct {
	Status  string      `json:"status" example:"success" description:"Response status (success or error)"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully" description:"Response message"`
	Data    interface{} `json:"data,omitempty" description:"Response data payload"`
	Error   string      `json:"error,omitempty" example:"Invalid input" description:"Error message (only present when status is error)"`
}

// SwaggerPaginatedResponse represents a paginated API response
// @Description A paginated API response format
type SwaggerPaginatedResponse struct {
	Items interface{} `json:"items" description:"Array of items for the current page"`
	Meta  struct {
		Page     int `json:"page" example:"1" description:"Current page number"`
		Limit    int `json:"limit" example:"10" description:"Number of items per page"`
		Total    int `json:"total" example:"50" description:"Total number of items"`
		LastPage int `json:"last_page" example:"5" description:"Last page number"`
	} `json:"meta" description:"Pagination metadata"`
}

// SwaggerPostsResponse represents the response for listing posts
// @Description Response model for listing blog posts
type SwaggerPostsResponse struct {
	Posts []Post `json:"posts" description:"List of posts"`
	Meta  struct {
		Page     int `json:"page" example:"1" description:"Current page number"`
		Limit    int `json:"limit" example:"10" description:"Number of items per page"`
		Total    int `json:"total" example:"50" description:"Total number of items"`
		LastPage int `json:"last_page" example:"5" description:"Last page number"`
	} `json:"meta" description:"Pagination metadata"`
}

// SwaggerProfileResponse represents the user profile response
// @Description Response model for user profile information
type SwaggerProfileResponse struct {
	ID           uint      `json:"id" example:"1" description:"User ID"`
	Username     string    `json:"username" example:"johndoe" description:"Username"`
	Email        string    `json:"email" example:"john@example.com" description:"Email address"`
	FirstName    string    `json:"first_name,omitempty" example:"John" description:"First name"`
	LastName     string    `json:"last_name,omitempty" example:"Doe" description:"Last name"`
	Bio          string    `json:"bio,omitempty" example:"Software developer" description:"User biography"`
	ProfileImage string    `json:"profile_image,omitempty" example:"https://example.com/avatar.jpg" description:"Profile image URL"`
	Role         string    `json:"role" example:"user" description:"User role"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" description:"Account creation timestamp"`
}

// SwaggerUpdateProfileRequest represents the request to update a user profile
// @Description Request model for updating user profile
type SwaggerUpdateProfileRequest struct {
	FirstName string `json:"first_name,omitempty" example:"John" description:"First name"`
	LastName  string `json:"last_name,omitempty" example:"Doe" description:"Last name"`
	Bio       string `json:"bio,omitempty" example:"Software developer" description:"User biography"`
}

// SwaggerAvatarResponse represents the response after uploading an avatar
// @Description Response model for avatar upload
type SwaggerAvatarResponse struct {
	ProfileImage string `json:"profile_image" example:"https://example.com/avatar.jpg" description:"URL to the uploaded avatar"`
}

// SwaggerPostCoverResponse represents the response after uploading a post cover
// @Description Response model for post cover upload
type SwaggerPostCoverResponse struct {
	Cover string `json:"cover" example:"https://example.com/cover.jpg" description:"URL to the uploaded cover image"`
}
