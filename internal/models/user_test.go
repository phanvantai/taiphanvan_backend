package models_test

import (
	"testing"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserModel(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the user model
	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)

	t.Run("Create user", func(t *testing.T) {
		// Create a new user
		user := models.User{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "hashedpassword", // In real-world scenarios, this would be hashed
			FirstName: "Test",
			LastName:  "User",
			Bio:       "This is a test user",
			Role:      "user",
		}

		// Save the user to the database
		result := db.Create(&user)
		assert.NoError(t, result.Error)
		assert.NotZero(t, user.ID, "User ID should not be zero after creation")
		assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")
	})

	t.Run("Find user by ID", func(t *testing.T) {
		// Create a user first
		user := models.User{
			Username:  "finduser",
			Email:     "find@example.com",
			Password:  "hashedpassword",
			FirstName: "Find",
			LastName:  "User",
			Role:      "user",
		}
		db.Create(&user)

		// Find the user by ID
		var foundUser models.User
		result := db.First(&foundUser, user.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Username, foundUser.Username)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("Update user", func(t *testing.T) {
		// Create a user first
		user := models.User{
			Username:  "updateuser",
			Email:     "update@example.com",
			Password:  "hashedpassword",
			FirstName: "Update",
			LastName:  "User",
			Role:      "user",
		}
		db.Create(&user)

		// Update the user
		originalUpdatedAt := user.UpdatedAt
		time.Sleep(1 * time.Millisecond) // Ensure time difference in UpdatedAt

		user.FirstName = "Updated"
		user.LastName = "Name"
		user.Bio = "Updated bio"

		result := db.Save(&user)
		assert.NoError(t, result.Error)

		// Verify the update
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.Equal(t, "Updated", updatedUser.FirstName)
		assert.Equal(t, "Name", updatedUser.LastName)
		assert.Equal(t, "Updated bio", updatedUser.Bio)
		assert.True(t, updatedUser.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
	})

	t.Run("Delete user", func(t *testing.T) {
		// Create a user first
		user := models.User{
			Username:  "deleteuser",
			Email:     "delete@example.com",
			Password:  "hashedpassword",
			FirstName: "Delete",
			LastName:  "User",
			Role:      "user",
		}
		db.Create(&user)

		// Delete the user
		result := db.Delete(&user)
		assert.NoError(t, result.Error)

		// Try to find the deleted user (should be soft-deleted)
		var deletedUser models.User
		result = db.First(&deletedUser, user.ID)
		assert.Error(t, result.Error, "Should not find soft-deleted user")

		// Find with unscoped (should find the soft-deleted user)
		result = db.Unscoped().First(&deletedUser, user.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, deletedUser.DeletedAt.Time)
	})

	t.Run("Unique email constraint", func(t *testing.T) {
		// Create a user first
		user1 := models.User{
			Username:  "uniqueuser1",
			Email:     "unique@example.com",
			Password:  "hashedpassword",
			FirstName: "Unique",
			LastName:  "User",
			Role:      "user",
		}
		result := db.Create(&user1)
		assert.NoError(t, result.Error)

		// Try to create another user with the same email
		user2 := models.User{
			Username:  "uniqueuser2",
			Email:     "unique@example.com", // Same email
			Password:  "hashedpassword",
			FirstName: "Unique",
			LastName:  "User2",
			Role:      "user",
		}
		result = db.Create(&user2)
		assert.Error(t, result.Error, "Should error on duplicate email")
	})

	t.Run("Unique username constraint", func(t *testing.T) {
		// Create a user first
		user1 := models.User{
			Username:  "sameusername",
			Email:     "username1@example.com",
			Password:  "hashedpassword",
			FirstName: "Username",
			LastName:  "Test",
			Role:      "user",
		}
		result := db.Create(&user1)
		assert.NoError(t, result.Error)

		// Try to create another user with the same username
		user2 := models.User{
			Username:  "sameusername", // Same username
			Email:     "username2@example.com",
			Password:  "hashedpassword",
			FirstName: "Username",
			LastName:  "Test2",
			Role:      "user",
		}
		result = db.Create(&user2)
		assert.Error(t, result.Error, "Should error on duplicate username")
	})
}
