// Package models provides data models for the API
package models

import "time"

// SwaggerDeletedAt is a custom type for Swagger documentation
// @Description A timestamp for soft-deleted records (null if not deleted)
type SwaggerDeletedAt struct {
	Time  time.Time `json:"time,omitempty"`
	Valid bool      `json:"valid,omitempty"`
}

// SwaggerStandardResponse represents a standard API response with generic data type
// @Description A standard API response format with type-safe data payload
type SwaggerStandardResponse[T any] struct {
	Status  string `json:"status" example:"success" description:"Response status (success or error)"`
	Message string `json:"message,omitempty" example:"Operation completed successfully" description:"Response message"`
	Data    T      `json:"data,omitempty" description:"Response data payload"`
	Error   string `json:"error,omitempty" example:"Invalid input" description:"Error message (only present when status is error)"`
}

// NewSuccessResponse creates a new success response with the given data and message
func NewSuccessResponse[T any](data T, message string) SwaggerStandardResponse[T] {
	return SwaggerStandardResponse[T]{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response with the given error message and error type
func NewErrorResponse(errorType, message string) SwaggerStandardResponse[any] {
	return SwaggerStandardResponse[any]{
		Status:  "error",
		Error:   errorType,
		Message: message,
	}
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

// SwaggerFileUploadResponse represents the response after uploading a file for editor use
// @Description Response model for editor file upload
type SwaggerFileUploadResponse struct {
	FileURL string `json:"file_url" example:"https://example.com/file.jpg" description:"URL to the uploaded file"`
}

// SwaggerDeleteFileRequest represents a request to delete a file
// @Description Request model for deleting a file
type SwaggerDeleteFileRequest struct {
	FileURL string `json:"file_url" example:"https://example.com/file.jpg" description:"URL of the file to delete"`
}

// SwaggerViewCountResponse represents the response for the view count increment endpoint
// @Description Response model for incrementing a post's view count
type SwaggerViewCountResponse struct {
	Status  string `json:"status" example:"success" description:"Response status (success or error)"`
	Message string `json:"message" example:"View count incremented successfully" description:"Response message"`
	Data    struct {
		PostID    uint `json:"post_id" example:"123" description:"ID of the post"`
		ViewCount uint `json:"view_count" example:"42" description:"New view count after incrementing"`
	} `json:"data" description:"Response data payload"`
}
