package models_test

import (
	"testing"
	"time"

	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenModels(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the necessary models
	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.BlacklistedToken{})
	assert.NoError(t, err)

	// Create a test user for tokens
	user := models.User{
		Username:  "tokenuser",
		Email:     "token@example.com",
		Password:  "hashedpassword",
		FirstName: "Token",
		LastName:  "User",
		Role:      "user",
	}
	db.Create(&user)

	t.Run("Create refresh token", func(t *testing.T) {
		// Create a new refresh token
		refreshToken := models.RefreshToken{
			Token:     "valid_refresh_token_string",
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
			Revoked:   false,
		}

		// Save the token to the database
		result := db.Create(&refreshToken)
		assert.NoError(t, result.Error)
		assert.NotZero(t, refreshToken.ID, "RefreshToken ID should not be zero after creation")
		assert.False(t, refreshToken.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.False(t, refreshToken.UpdatedAt.IsZero(), "UpdatedAt should be set")
	})

	t.Run("Find refresh token", func(t *testing.T) {
		// Create a token first
		tokenString := "find_refresh_token_string"
		refreshToken := models.RefreshToken{
			Token:     tokenString,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
			Revoked:   false,
		}
		db.Create(&refreshToken)

		// Find the token
		var foundToken models.RefreshToken
		result := db.Where("token = ?", tokenString).First(&foundToken)
		assert.NoError(t, result.Error)
		assert.Equal(t, tokenString, foundToken.Token)
		assert.Equal(t, user.ID, foundToken.UserID)
		assert.False(t, foundToken.Revoked)
	})

	t.Run("Update refresh token revoked status", func(t *testing.T) {
		// Create a token first
		tokenString := "revoke_refresh_token_string"
		refreshToken := models.RefreshToken{
			Token:     tokenString,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
			Revoked:   false,
		}
		db.Create(&refreshToken)

		// Update the token revoked status
		result := db.Model(&refreshToken).Update("revoked", true)
		assert.NoError(t, result.Error)

		// Verify the update
		var updatedToken models.RefreshToken
		db.Where("token = ?", tokenString).First(&updatedToken)
		assert.True(t, updatedToken.Revoked)
	})

	t.Run("Create blacklisted token", func(t *testing.T) {
		// Create a new blacklisted token
		blacklistedToken := models.BlacklistedToken{
			Token:     "revoked_access_token_string",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		// Save the token to the database
		result := db.Create(&blacklistedToken)
		assert.NoError(t, result.Error)
		assert.NotZero(t, blacklistedToken.ID, "BlacklistedToken ID should not be zero after creation")
		assert.False(t, blacklistedToken.CreatedAt.IsZero(), "CreatedAt should be set")
	})

	t.Run("Check for blacklisted token", func(t *testing.T) {
		// Create a blacklisted token first
		tokenString := "check_blacklisted_token_string"
		blacklistedToken := models.BlacklistedToken{
			Token:     tokenString,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		db.Create(&blacklistedToken)

		// Check if the token is blacklisted
		var count int64
		result := db.Model(&models.BlacklistedToken{}).Where("token = ?", tokenString).Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), count, "Should find the blacklisted token")
	})

	t.Run("Unique refresh token constraint", func(t *testing.T) {
		// Create a token first
		tokenString := "unique_refresh_token_string"
		refreshToken1 := models.RefreshToken{
			Token:     tokenString,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
			Revoked:   false,
		}
		result := db.Create(&refreshToken1)
		assert.NoError(t, result.Error)

		// Try to create another token with the same string
		refreshToken2 := models.RefreshToken{
			Token:     tokenString, // Same token string
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
			Revoked:   false,
		}
		result = db.Create(&refreshToken2)
		assert.Error(t, result.Error, "Should error on duplicate token string")
	})

	t.Run("Unique blacklisted token constraint", func(t *testing.T) {
		// Create a blacklisted token first
		tokenString := "unique_blacklisted_token_string"
		blacklistedToken1 := models.BlacklistedToken{
			Token:     tokenString,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		result := db.Create(&blacklistedToken1)
		assert.NoError(t, result.Error)

		// Try to create another token with the same string
		blacklistedToken2 := models.BlacklistedToken{
			Token:     tokenString, // Same token string
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		result = db.Create(&blacklistedToken2)
		assert.Error(t, result.Error, "Should error on duplicate token string")
	})
}
