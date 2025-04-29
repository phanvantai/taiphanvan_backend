package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a blog user
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"size:50;not null;unique"`
	Email        string         `json:"email" gorm:"size:100;not null;unique"`
	Password     string         `json:"-" gorm:"size:100;not null"` // Password is not included in JSON responses
	FirstName    string         `json:"first_name" gorm:"size:50"`
	LastName     string         `json:"last_name" gorm:"size:50"`
	Bio          string         `json:"bio" gorm:"type:text"`
	Role         string         `json:"role" gorm:"size:20;default:'user'"` // admin, editor, user
	ProfileImage string         `json:"profile_image" gorm:"size:255"`
	Posts        []Post         `json:"posts" gorm:"foreignKey:UserID"`
	Comments     []Comment      `json:"comments" gorm:"foreignKey:UserID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Token response structure
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// LoginRequest structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest structure
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
