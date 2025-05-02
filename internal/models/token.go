package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	gorm.Model
	Token     string    `gorm:"size:255;not null;uniqueIndex"`
	UserID    uint      `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	IssuedAt  time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
	// User is used for creating "belongs to" relationship
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

// BlacklistedToken represents a revoked JWT token
type BlacklistedToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Token     string         `json:"token" gorm:"size:500;not null;uniqueIndex"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// TokenResponse represents the response after successful authentication
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // in seconds
}

// RefreshTokenRequest represents a request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenRevokeRequest represents a request to revoke a refresh token
type TokenRevokeRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
