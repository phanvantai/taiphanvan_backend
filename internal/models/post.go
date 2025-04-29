package models

import (
	"time"

	"gorm.io/gorm"
)

// Post represents a blog post
type Post struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Slug        string         `json:"slug" gorm:"size:255;not null;unique"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Excerpt     string         `json:"excerpt" gorm:"type:text"`
	UserID      uint           `json:"user_id"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Tags        []Tag          `json:"tags" gorm:"many2many:post_tags;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	PublishedAt *time.Time     `json:"published_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Tag represents a post tag
type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"size:50;not null;unique"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags;"`
}

// Comment represents a user comment on a post
type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	UserID    uint           `json:"user_id"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	PostID    uint           `json:"post_id"`
	Post      Post           `json:"post" gorm:"foreignKey:PostID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
