package models

import (
	"time"

	"gorm.io/gorm"
)

// Post represents a blog post
// @Description A blog post with content, metadata, and relationships
type Post struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Title       string         `json:"title" gorm:"size:255;not null" example:"My First Blog Post" description:"Post title"`
	Slug        string         `json:"slug" gorm:"size:255;not null;unique" example:"my-first-blog-post" description:"URL-friendly version of the title"`
	Content     string         `json:"content" gorm:"type:text;not null" example:"This is the content of my blog post..." description:"Main content of the post"`
	Excerpt     string         `json:"excerpt" gorm:"type:text" example:"A short summary of the post" description:"Short summary or preview of the post"`
	Cover       string         `json:"cover" gorm:"size:500" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg" description:"URL to the post's cover image"`
	UserID      uint           `json:"user_id" example:"1" description:"ID of the post author"`
	User        User           `json:"user" gorm:"foreignKey:UserID" description:"Author of the post"`
	Tags        []Tag          `json:"tags" gorm:"many2many:post_tags;" description:"Tags associated with the post"`
	CreatedAt   time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the post was created"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the post was last updated"`
	PublishedAt *time.Time     `json:"published_at" example:"2023-01-03T12:00:00Z" description:"When the post was published (null if draft)"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
}

// Tag represents a post tag
// @Description A tag that can be associated with multiple posts
type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Name  string `json:"name" gorm:"size:50;not null;unique" example:"technology" description:"Tag name"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags;" description:"Posts associated with this tag"`
}

// Comment represents a user comment on a post
// @Description A comment made by a user on a specific post
type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Content   string         `json:"content" gorm:"type:text;not null" example:"Great post!" description:"Comment content"`
	UserID    uint           `json:"user_id" example:"1" description:"ID of the comment author"`
	User      User           `json:"user" gorm:"foreignKey:UserID" description:"Author of the comment"`
	PostID    uint           `json:"post_id" example:"1" description:"ID of the post being commented on"`
	Post      Post           `json:"post" gorm:"foreignKey:PostID" description:"Post being commented on"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the comment was created"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the comment was last updated"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
}
