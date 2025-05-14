package models

import (
	"time"

	"gorm.io/gorm"
)

// PostStatus represents the publication status of a post
type PostStatus string

const (
	// PostStatusDraft indicates the post is a draft and not publicly visible
	PostStatusDraft PostStatus = "draft"
	// PostStatusPublished indicates the post is published and publicly visible
	PostStatusPublished PostStatus = "published"
	// PostStatusArchived indicates the post has been archived
	PostStatusArchived PostStatus = "archived"
	// PostStatusScheduled indicates the post is scheduled to be published at a future date
	PostStatusScheduled PostStatus = "scheduled"
)

// Post represents a blog post
// @Description A blog post with content, metadata, and relationships
type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Title     string         `json:"title" gorm:"size:255;not null" example:"My First Blog Post" description:"Post title"`
	Slug      string         `json:"slug" gorm:"size:255;not null;unique" example:"my-first-blog-post" description:"URL-friendly version of the title"`
	Content   string         `json:"content" gorm:"type:text;not null" example:"This is the content of my blog post..." description:"Main content of the post"`
	Excerpt   string         `json:"excerpt" gorm:"type:text" example:"A short summary of the post" description:"Short summary or preview of the post"`
	Cover     string         `json:"cover" gorm:"size:500" example:"https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg" description:"URL to the post's cover image"`
	Status    PostStatus     `json:"status" gorm:"type:varchar(20);not null;default:'draft'" example:"published" description:"Publication status of the post"`
	ViewCount uint           `json:"view_count" gorm:"default:0" example:"42" description:"Number of times the post has been viewed"`
	UserID    uint           `json:"user_id" example:"1" description:"ID of the post author"`
	User      User           `json:"user" gorm:"foreignKey:UserID" description:"Author of the post"`
	Tags      []Tag          `json:"tags" gorm:"many2many:post_tags;" description:"Tags associated with the post"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the post was created"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the post was last updated"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
}

// Tag represents a post tag
// @Description A tag that can be associated with multiple posts
type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Name  string `json:"name" gorm:"size:50;not null;unique" example:"technology" description:"Tag name"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags;" description:"Posts associated with this tag"`
}

// VoteType represents the type of vote (upvote or downvote)
type VoteType int8

const (
	// VoteTypeDownvote represents a downvote
	VoteTypeDownvote VoteType = -1
	// VoteTypeNone represents no vote
	VoteTypeNone VoteType = 0
	// VoteTypeUpvote represents an upvote
	VoteTypeUpvote VoteType = 1
)

// CommentVote represents a user's vote on a comment
// @Description A vote (upvote or downvote) by a user on a comment
type CommentVote struct {
	ID        uint      `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	UserID    uint      `json:"user_id" example:"1" description:"ID of the user who voted"`
	CommentID uint      `json:"comment_id" example:"1" description:"ID of the comment that was voted on"`
	VoteType  VoteType  `json:"vote_type" gorm:"type:smallint;not null" example:"1" description:"Type of vote: 1 (upvote), -1 (downvote), 0 (none)"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the vote was created"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the vote was last updated"`
}

// Comment represents a user comment on a post
// @Description A comment made by a user on a specific post
type Comment struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1" description:"Unique identifier"`
	Content     string         `json:"content" gorm:"type:text;not null" example:"Great post!" description:"Comment content"`
	UserID      uint           `json:"user_id" example:"1" description:"ID of the comment author"`
	User        User           `json:"user" gorm:"foreignKey:UserID" description:"Author of the comment"`
	PostID      uint           `json:"post_id" example:"1" description:"ID of the post being commented on"`
	Post        Post           `json:"post" gorm:"foreignKey:PostID" description:"Post being commented on"`
	UpvoteCount int            `json:"upvote_count" gorm:"default:0" example:"5" description:"Number of upvotes on the comment"`
	Votes       []CommentVote  `json:"-" gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE;" description:"Users who voted on this comment"`
	CreatedAt   time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z" description:"When the comment was created"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2023-01-02T12:00:00Z" description:"When the comment was last updated"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"` // Hide from Swagger
}

// CreatePostRequest represents the request body for creating a new post
// @Description Request model for creating a new blog post
type CreatePostRequest struct {
	Title     string     `json:"title" binding:"required" example:"My New Post" description:"Post title"`
	Content   string     `json:"content" binding:"required" example:"This is the content of my new post" description:"Main content of the post"`
	Excerpt   string     `json:"excerpt" example:"A short excerpt" description:"Short summary or preview of the post"`
	Cover     string     `json:"cover" example:"https://example.com/image.jpg" description:"URL to the post's cover image"`
	Tags      []string   `json:"tags" example:"[\"technology\",\"programming\"]" description:"Tags associated with the post"`
	Status    PostStatus `json:"status" example:"published" description:"Publication status of the post (draft, published, archived, scheduled)"`
	PublishAt *time.Time `json:"publish_at,omitempty" example:"2023-01-03T12:00:00Z" description:"When to publish the post if status is 'scheduled'"`
}

// UpdatePostRequest represents the request body for updating an existing post
// @Description Request model for updating an existing blog post
type UpdatePostRequest struct {
	Title     *string     `json:"title" example:"Updated Post Title" description:"New post title"`
	Content   *string     `json:"content" example:"Updated content" description:"New main content of the post"`
	Excerpt   *string     `json:"excerpt" example:"Updated excerpt" description:"New short summary or preview of the post"`
	Cover     *string     `json:"cover" example:"https://example.com/updated-cover.jpg" description:"New URL to the post's cover image"`
	Tags      []string    `json:"tags" example:"[\"technology\",\"programming\",\"updated\"]" description:"New tags associated with the post"`
	Status    *PostStatus `json:"status" example:"published" description:"New publication status of the post"`
	PublishAt *time.Time  `json:"publish_at,omitempty" example:"2023-01-03T12:00:00Z" description:"When to publish the post if status is 'scheduled'"`
}

// CreateCommentRequest represents the request body for creating a new comment
// @Description Request model for creating a new comment on a post
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required" example:"This is a great post!" description:"Comment content"`
}

// UpdateCommentRequest represents the request body for updating an existing comment
// @Description Request model for updating an existing comment
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required" example:"This is my updated comment" description:"Updated comment content"`
}

// CommentVoteRequest represents the request body for voting on a comment
// @Description Request model for voting on a comment
type CommentVoteRequest struct {
	VoteType VoteType `json:"vote_type" binding:"required" example:"1" description:"Type of vote: 1 (upvote), -1 (downvote), 0 (none)"`
}

// CommentVoteResponse represents the response body for a comment vote operation
// @Description Response model for a comment vote operation
type CommentVoteResponse struct {
	CommentID   uint `json:"comment_id" example:"1" description:"ID of the comment that was voted on"`
	UpvoteCount int  `json:"upvote_count" example:"5" description:"Updated upvote count for the comment"`
	UserVote    int8 `json:"user_vote" example:"1" description:"The user's current vote: 1 (upvote), -1 (downvote), 0 (none)"`
}

// CommentWithUserVote represents a comment with the current user's vote
// @Description A comment with additional information about the current user's vote
type CommentWithUserVote struct {
	Comment  Comment `json:"comment" description:"The comment data"`
	UserVote int8    `json:"user_vote" example:"1" description:"The user's current vote: 1 (upvote), -1 (downvote), 0 (none)"`
}

// TagWithCount represents a tag with its post count
// @Description A tag with the count of posts using it
type TagWithCount struct {
	ID        uint   `json:"id" example:"1" description:"Unique identifier"`
	Name      string `json:"name" example:"technology" description:"Tag name"`
	PostCount int64  `json:"post_count" example:"5" description:"Number of posts using this tag"`
}

// SetPostStatusRequest represents the request body for updating a post's status
// @Description Request model for changing a post's status
type SetPostStatusRequest struct {
	Status    PostStatus `json:"status" binding:"required" example:"published" description:"New post status (draft, published, archived, scheduled)"`
	PublishAt *time.Time `json:"publish_at,omitempty" example:"2023-01-03T12:00:00Z" description:"When to publish the post if status is 'scheduled'"`
}
