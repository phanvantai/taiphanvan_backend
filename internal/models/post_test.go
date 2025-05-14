package models_test

import (
	"testing"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPostModel(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the necessary models
	err = db.AutoMigrate(&models.User{}, &models.Post{}, &models.Tag{})
	assert.NoError(t, err)

	// Create a test user for posts
	user := models.User{
		Username:  "postuser",
		Email:     "post@example.com",
		Password:  "hashedpassword",
		FirstName: "Post",
		LastName:  "User",
		Role:      "user",
	}
	db.Create(&user)

	t.Run("Create post", func(t *testing.T) {
		// Create a new post
		post := models.Post{
			Title:    "Test Post",
			Slug:     "test-post",
			Content:  "This is a test post content",
			Excerpt:  "Test excerpt",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}

		// Save the post to the database
		result := db.Create(&post)
		assert.NoError(t, result.Error)
		assert.NotZero(t, post.ID, "Post ID should not be zero after creation")
		assert.False(t, post.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.False(t, post.UpdatedAt.IsZero(), "UpdatedAt should be set")
		assert.Equal(t, models.PostStatusDraft, post.Status)
		assert.Equal(t, user.ID, post.UserID)
	})

	t.Run("Find post by ID", func(t *testing.T) {
		// Create a post first
		post := models.Post{
			Title:    "Find Post",
			Slug:     "find-post",
			Content:  "This is a post to be found",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		db.Create(&post)

		// Find the post by ID
		var foundPost models.Post
		result := db.First(&foundPost, post.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, post.ID, foundPost.ID)
		assert.Equal(t, post.Title, foundPost.Title)
		assert.Equal(t, post.Slug, foundPost.Slug)
		assert.Equal(t, post.Content, foundPost.Content)
	})

	t.Run("Update post", func(t *testing.T) {
		// Create a post first
		post := models.Post{
			Title:    "Update Post",
			Slug:     "update-post",
			Content:  "Original content",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		db.Create(&post)

		// Update the post
		originalUpdatedAt := post.UpdatedAt
		time.Sleep(1 * time.Millisecond) // Ensure time difference in UpdatedAt

		post.Title = "Updated Title"
		post.Content = "Updated content"
		post.Status = models.PostStatusPublished
		
		result := db.Save(&post)
		assert.NoError(t, result.Error)

		// Verify the update
		var updatedPost models.Post
		db.First(&updatedPost, post.ID)
		assert.Equal(t, "Updated Title", updatedPost.Title)
		assert.Equal(t, "Updated content", updatedPost.Content)
		assert.Equal(t, models.PostStatusPublished, updatedPost.Status)
		assert.True(t, updatedPost.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
	})

	t.Run("Delete post", func(t *testing.T) {
		// Create a post first
		post := models.Post{
			Title:    "Delete Post",
			Slug:     "delete-post",
			Content:  "This post will be deleted",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		db.Create(&post)

		// Delete the post
		result := db.Delete(&post)
		assert.NoError(t, result.Error)

		// Try to find the deleted post (should be soft-deleted)
		var deletedPost models.Post
		result = db.First(&deletedPost, post.ID)
		assert.Error(t, result.Error, "Should not find soft-deleted post")
		
		// Find with unscoped (should find the soft-deleted post)
		result = db.Unscoped().First(&deletedPost, post.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, deletedPost.DeletedAt.Time)
	})

	t.Run("Unique slug constraint", func(t *testing.T) {
		// Create a post first
		post1 := models.Post{
			Title:    "Unique Slug Post 1",
			Slug:     "unique-slug",
			Content:  "First post with this slug",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		result := db.Create(&post1)
		assert.NoError(t, result.Error)

		// Try to create another post with the same slug
		post2 := models.Post{
			Title:    "Unique Slug Post 2",
			Slug:     "unique-slug", // Same slug
			Content:  "Second post with this slug",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		result = db.Create(&post2)
		assert.Error(t, result.Error, "Should error on duplicate slug")
	})

	t.Run("Post with tags", func(t *testing.T) {
		// Create some tags first
		tag1 := models.Tag{Name: "Technology"}
		tag2 := models.Tag{Name: "Programming"}
		db.Create(&tag1)
		db.Create(&tag2)

		// Create a post with tags
		post := models.Post{
			Title:    "Post with Tags",
			Slug:     "post-with-tags",
			Content:  "This post has multiple tags",
			Status:   models.PostStatusPublished,
			UserID:   user.ID,
			Tags:     []models.Tag{tag1, tag2},
		}
		result := db.Create(&post)
		assert.NoError(t, result.Error)

		// Find the post with tags
		var foundPost models.Post
		result = db.Preload("Tags").First(&foundPost, post.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, 2, len(foundPost.Tags))
		
		// Verify the tags
		tagNames := []string{foundPost.Tags[0].Name, foundPost.Tags[1].Name}
		assert.Contains(t, tagNames, "Technology")
		assert.Contains(t, tagNames, "Programming")
	})

	t.Run("Find posts by user", func(t *testing.T) {
		// Create another user
		anotherUser := models.User{
			Username:  "anotheruser",
			Email:     "another@example.com",
			Password:  "hashedpassword",
			FirstName: "Another",
			LastName:  "User",
			Role:      "user",
		}
		db.Create(&anotherUser)

		// Create posts for both users
		post1 := models.Post{
			Title:    "User 1 Post",
			Slug:     "user-1-post",
			Content:  "Post by first user",
			Status:   models.PostStatusPublished,
			UserID:   user.ID,
		}
		post2 := models.Post{
			Title:    "User 2 Post",
			Slug:     "user-2-post",
			Content:  "Post by second user",
			Status:   models.PostStatusPublished,
			UserID:   anotherUser.ID,
		}
		db.Create(&post1)
		db.Create(&post2)

		// Find posts by the first user
		var posts []models.Post
		result := db.Where("user_id = ?", user.ID).Find(&posts)
		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, len(posts), 1, "Should find at least one post by first user")
		
		// Check that all returned posts belong to the first user
		for _, p := range posts {
			assert.Equal(t, user.ID, p.UserID)
		}
	})

	t.Run("Find published posts", func(t *testing.T) {
		// Create posts with different statuses
		publishedPost := models.Post{
			Title:    "Published Post",
			Slug:     "published-post",
			Content:  "This post is published",
			Status:   models.PostStatusPublished,
			UserID:   user.ID,
		}
		draftPost := models.Post{
			Title:    "Draft Post",
			Slug:     "draft-post",
			Content:  "This post is a draft",
			Status:   models.PostStatusDraft,
			UserID:   user.ID,
		}
		db.Create(&publishedPost)
		db.Create(&draftPost)

		// Find only published posts
		var posts []models.Post
		result := db.Where("status = ?", models.PostStatusPublished).Find(&posts)
		assert.NoError(t, result.Error)
		
		// Check that all returned posts are published
		for _, p := range posts {
			assert.Equal(t, models.PostStatusPublished, p.Status)
		}
	})
}
