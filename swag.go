// Package main provides type aliases for Swagger documentation
package main

import (
	"github.com/phanvantai/taiphanvan_backend/internal/models"
)

// These type aliases are used to simplify the model names in Swagger documentation
type (
	// Auth models
	LoginRequest        = models.LoginRequest
	RegisterRequest     = models.RegisterRequest
	RefreshTokenRequest = models.RefreshTokenRequest
	TokenRevokeRequest  = models.TokenRevokeRequest
	TokenResponse       = models.TokenResponse

	// User models
	UpdateProfileRequest = models.UpdateProfileRequest
	AvatarResponse       = models.AvatarResponse

	// Comment models
	CreateCommentRequest = models.CreateCommentRequest
	UpdateCommentRequest = models.UpdateCommentRequest
	CommentVoteRequest   = models.CommentVoteRequest
	CommentVoteResponse  = models.CommentVoteResponse

	// Post models
	CreatePostRequest    = models.CreatePostRequest
	UpdatePostRequest    = models.UpdatePostRequest
	SetPostStatusRequest = models.SetPostStatusRequest
	PostsResponse        = models.PostsResponse
	PostCoverResponse    = models.PostCoverResponse

	// File models
	FileUploadResponse = models.FileUploadResponse
	DeleteFileRequest  = models.DeleteFileRequest

	// Common models
	StandardResponse         = models.StandardResponse
	StandardResponseUser     = models.StandardResponseUser
	StandardResponseString   = models.StandardResponseString
	SwaggerViewCountResponse = models.ViewCountResponse
	ViewCountResponse        = models.ViewCountResponse
	PaginationMeta           = models.PaginationMeta
)
