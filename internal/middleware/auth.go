package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
)

// JWT claims structure
type Claims struct {
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AppConfig holds application configuration
var AppConfig *config.Config

// SetConfig sets the application configuration for middleware
func SetConfig(cfg *config.Config) {
	AppConfig = cfg
}

// AuthMiddleware checks for valid JWT access token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"error":   "Authorization required",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		var count int64
		if err := database.DB.Model(&models.BlacklistedToken{}).
			Where("token = ?", tokenString).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"error":   "Authentication error",
				"message": "Failed to validate token",
			})
			c.Abort()
			return
		}

		if count > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"error":   "Token revoked",
				"message": "This token has been revoked. Please log in again",
			})
			c.Abort()
			return
		}

		// Parse token
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Ensure this is an access token, not a refresh token
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"error":   "Invalid token type",
				"message": "Refresh token cannot be used for authentication",
			})
			c.Abort()
			return
		}

		// Set the user ID in the context for later use
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"error":   "Permission denied",
				"message": "Admin privileges required for this resource",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GenerateTokenPair creates both access and refresh tokens for a user
func GenerateTokenPair(user models.User) (accessToken string, refreshToken string, refreshTokenID uint, err error) {
	if AppConfig == nil {
		return "", "", 0, errors.New("application configuration not set")
	}

	// Generate access token
	accessToken, err = generateAccessToken(user)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, refreshTokenModel, err := generateRefreshToken(user)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, refreshTokenModel.ID, nil
}

// generateAccessToken creates a new JWT access token for a user
func generateAccessToken(user models.User) (string, error) {
	if AppConfig == nil {
		return "", errors.New("application configuration not set")
	}

	// Set expiration time based on config
	expirationTime := time.Now().Add(AppConfig.JWT.AccessExpiry)

	// Create the claims with user information
	claims := &Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create the token with claims and sign it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(AppConfig.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// generateRefreshToken creates a new JWT refresh token and stores it in the database
func generateRefreshToken(user models.User) (string, *models.RefreshToken, error) {
	if AppConfig == nil {
		return "", nil, errors.New("application configuration not set")
	}

	// Set expiration time for refresh token based on config
	issuedAt := time.Now()
	expirationTime := issuedAt.Add(AppConfig.JWT.RefreshExpiry)

	// Create the claims with user information
	claims := &Claims{
		UserID:    user.ID,
		Role:      user.Role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create the token with claims and sign it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(AppConfig.JWT.Secret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token in database
	refreshToken := &models.RefreshToken{
		Token:     tokenString,
		UserID:    user.ID,
		ExpiresAt: expirationTime,
		IssuedAt:  issuedAt,
		Revoked:   false,
	}

	if result := database.DB.Create(refreshToken); result.Error != nil {
		return "", nil, fmt.Errorf("failed to store refresh token: %w", result.Error)
	}

	return tokenString, refreshToken, nil
}

// RefreshAccessToken creates a new access token from a valid refresh token
func RefreshAccessToken(refreshToken string) (accessToken string, err error) {
	// Validate the refresh token
	claims, err := validateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure this is a refresh token
	if claims.TokenType != "refresh" {
		return "", errors.New("token is not a refresh token")
	}

	// Check if refresh token exists and is not revoked
	var dbToken models.RefreshToken
	if result := database.DB.Where("token = ? AND revoked = ?", refreshToken, false).First(&dbToken); result.Error != nil {
		return "", errors.New("refresh token has been revoked or does not exist")
	}

	// Check if token is expired
	if time.Now().After(dbToken.ExpiresAt) {
		return "", errors.New("refresh token has expired")
	}

	// Get user information to generate a new access token
	var user models.User
	if result := database.DB.First(&user, dbToken.UserID); result.Error != nil {
		return "", errors.New("user not found")
	}

	// Generate new access token
	newAccessToken, err := generateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked in the database
func RevokeRefreshToken(refreshToken string) error {
	var token models.RefreshToken
	if result := database.DB.Where("token = ?", refreshToken).First(&token); result.Error != nil {
		return errors.New("refresh token not found")
	}

	token.Revoked = true
	if result := database.DB.Save(&token); result.Error != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", result.Error)
	}

	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func RevokeAllUserRefreshTokens(userID uint) error {
	if result := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true); result.Error != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", result.Error)
	}

	return nil
}

// extractToken gets the token from the Authorization header
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// validateToken validates a token and returns its claims
func validateToken(tokenString string) (*Claims, error) {
	if AppConfig == nil {
		return nil, errors.New("application configuration not set")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
