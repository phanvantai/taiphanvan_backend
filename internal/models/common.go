// Package models provides data models for the API
package models

import "time"

type DeletedAt struct {
	Time  time.Time `json:"time,omitempty"`
	Valid bool      `json:"valid,omitempty"`
}

type StandardResponse[T any] struct {
	Status  string `json:"status" example:"success" description:"Response status (success or error)"`
	Message string `json:"message,omitempty" example:"Operation completed successfully" description:"Response message"`
	Data    T      `json:"data,omitempty" description:"Response data payload"`
	Error   string `json:"error,omitempty" example:"Invalid input" description:"Error message (only present when status is error)"`
}

type StandardResponseString struct {
	Status  string `json:"status" example:"success" description:"Response status (success or error)"`
	Message string `json:"message,omitempty" example:"Operation completed successfully" description:"Response message"`
	Data    string `json:"data,omitempty" example:"Some string data" description:"String data payload"`
	Error   string `json:"error,omitempty" example:"Invalid input" description:"Error message (only present when status is error)"`
}

type StandardResponseUser struct {
	Status  string         `json:"status" example:"success" description:"Response status (success or error)"`
	Message string         `json:"message,omitempty" example:"Operation completed successfully" description:"Response message"`
	Data    SwaggerProfile `json:"data,omitempty" description:"User profile data"`
	Error   string         `json:"error,omitempty" example:"Invalid input" description:"Error message (only present when status is error)"`
}

type SwaggerProfile struct {
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

// NewSuccessResponse creates a new success response with the given data and message
func NewSuccessResponse[T any](data T, message string) StandardResponse[T] {
	return StandardResponse[T]{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response with the given error message and error type
func NewErrorResponse(errorType, message string) StandardResponse[any] {
	return StandardResponse[any]{
		Status:  "error",
		Error:   errorType,
		Message: message,
	}
}

// PaginationMeta represents pagination metadata
// @Description Pagination metadata
type PaginationMeta struct {
	Page     int `json:"page" example:"1" description:"Current page number"`
	Limit    int `json:"limit" example:"10" description:"Number of items per page"`
	Total    int `json:"total" example:"50" description:"Total number of items"`
	LastPage int `json:"last_page" example:"5" description:"Last page number"`
}

// PaginatedResponse represents a paginated API response
// @Description A paginated API response format
type PaginatedResponse struct {
	Items interface{}    `json:"items" description:"Array of items for the current page"`
	Meta  PaginationMeta `json:"meta" description:"Pagination metadata"`
}

// PostsResponse represents the response for listing posts
// @Description Response model for listing blog posts
type PostsResponse struct {
	Posts []Post         `json:"posts" description:"List of posts"`
	Meta  PaginationMeta `json:"meta" description:"Pagination metadata"`
}

// UpdateProfileRequest represents the request to update a user profile
// @Description Request model for updating user profile
type UpdateProfileRequest struct {
	FirstName string `json:"first_name,omitempty" example:"John" description:"First name"`
	LastName  string `json:"last_name,omitempty" example:"Doe" description:"Last name"`
	Bio       string `json:"bio,omitempty" example:"Software developer" description:"User biography"`
}

// AvatarResponse represents the response after uploading an avatar
// @Description Response model for avatar upload
type AvatarResponse struct {
	ProfileImage string `json:"profile_image" example:"https://example.com/avatar.jpg" description:"URL to the uploaded avatar"`
}

// PostCoverResponse represents the response after uploading a post cover
// @Description Response model for post cover upload
type PostCoverResponse struct {
	Cover string `json:"cover" example:"https://example.com/cover.jpg" description:"URL to the uploaded cover image"`
}

// FileUploadResponse represents the response after uploading a file for editor use
// @Description Response model for editor file upload
type FileUploadResponse struct {
	FileURL string `json:"file_url" example:"https://example.com/file.jpg" description:"URL to the uploaded file"`
}

// DeleteFileRequest represents a request to delete a file
// @Description Request model for deleting a file
type DeleteFileRequest struct {
	FileURL string `json:"file_url" example:"https://example.com/file.jpg" description:"URL of the file to delete"`
}

// ViewCountResponse represents the response for the view count increment endpoint
// @Description Response model for incrementing a post's view count
type ViewCountResponse struct {
	Status  string `json:"status" example:"success" description:"Response status (success or error)"`
	Message string `json:"message" example:"View count incremented successfully" description:"Response message"`
	Data    struct {
		PostID    uint `json:"post_id" example:"123" description:"ID of the post"`
		ViewCount uint `json:"view_count" example:"42" description:"New view count after incrementing"`
	} `json:"data" description:"Response data payload"`
}
