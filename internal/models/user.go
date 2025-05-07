package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a blog user
// @Description A user account with profile information and relationships
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Username     string         `json:"username" gorm:"size:50;not null;unique" example:"johndoe" description:"Unique username"`
	Email        string         `json:"email" gorm:"size:100;not null;unique" example:"john@example.com" description:"Email address"`
	Password     string         `json:"-" gorm:"size:100;not null"` // Password is not included in JSON responses
	FirstName    string         `json:"first_name" gorm:"size:50" example:"John" description:"First name"`
	LastName     string         `json:"last_name" gorm:"size:50" example:"Doe" description:"Last name"`
	Bio          string         `json:"bio" gorm:"type:text" example:"I'm a software developer interested in web technologies." description:"User biography"`
	Role         string         `json:"role" gorm:"size:20;default:'user'" example:"user" description:"User role (admin, editor, user)"`
	ProfileImage string         `json:"profile_image" gorm:"size:255" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg" description:"URL to profile image"`
	Posts        []Post         `json:"posts,omitempty" gorm:"foreignKey:UserID" description:"Posts created by this user"`
	Comments     []Comment      `json:"comments,omitempty" gorm:"foreignKey:UserID" description:"Comments made by this user"`
	CreatedAt    time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the user account was created"`
	UpdatedAt    time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the user account was last updated"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
}
