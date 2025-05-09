package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/phanvantai/taiphanvan_backend/internal/config"
	"github.com/rs/zerolog/log"
)

// CloudinaryService handles interactions with Cloudinary
type CloudinaryService struct {
	cld *cloudinary.Cloudinary
	cfg config.CloudinaryConfig
}

const (
	// Folder paths for different upload types
	avatarFolder    = "avatars"
	postCoverFolder = "post_covers"
)

// NewCloudinaryService creates a new Cloudinary service
func NewCloudinaryService(cfg config.CloudinaryConfig) (*CloudinaryService, error) {
	if cfg.CloudName == "" || cfg.APIKey == "" || cfg.APISecret == "" {
		return nil, fmt.Errorf("missing Cloudinary configuration")
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return &CloudinaryService{
		cld: cld,
		cfg: cfg,
	}, nil
}

// UploadAvatar uploads an avatar image to Cloudinary
func (s *CloudinaryService) UploadAvatar(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create a unique public ID for the image
	timestamp := time.Now().UnixNano()
	folderPath := fmt.Sprintf("%s/%s", s.cfg.UploadFolder, avatarFolder)
	publicID := fmt.Sprintf("user_%d_%d", userID, timestamp)

	// Upload the file to Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "image",
		Folder:       folderPath,
	}

	log.Info().
		Str("public_id", publicID).
		Str("folder", folderPath).
		Uint("user_id", userID).
		Str("filename", file.Filename).
		Msg("Uploading avatar to Cloudinary")

	result, err := s.cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	log.Info().
		Str("public_id", publicID).
		Str("url", result.SecureURL).
		Uint("user_id", userID).
		Msg("Avatar uploaded successfully")

	return result.SecureURL, nil
}

// DeleteAvatar deletes an avatar from Cloudinary
func (s *CloudinaryService) DeleteAvatar(ctx context.Context, publicID string) error {
	if publicID == "" {
		return nil // Nothing to delete
	}

	log.Info().Str("public_id", publicID).Msg("Deleting avatar from Cloudinary")

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete from Cloudinary: %w", err)
	}

	log.Info().Str("public_id", publicID).Msg("Avatar deleted successfully")
	return nil
}

// UploadPostCover uploads a cover image for a post to Cloudinary
func (s *CloudinaryService) UploadPostCover(ctx context.Context, file *multipart.FileHeader, postID uint) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create a unique public ID for the image
	timestamp := time.Now().UnixNano()
	folderPath := fmt.Sprintf("%s/%s", s.cfg.UploadFolder, postCoverFolder)
	publicID := fmt.Sprintf("post_%d_%d", postID, timestamp)

	// Upload the file to Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "image",
		Folder:       folderPath,
	}

	log.Info().
		Str("public_id", publicID).
		Str("folder", folderPath).
		Uint("post_id", postID).
		Str("filename", file.Filename).
		Msg("Uploading post cover to Cloudinary")

	result, err := s.cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	log.Info().
		Str("public_id", publicID).
		Str("url", result.SecureURL).
		Uint("post_id", postID).
		Msg("Post cover uploaded successfully")

	return result.SecureURL, nil
}

// DeleteImage deletes an image from Cloudinary by URL
func (s *CloudinaryService) DeleteImage(ctx context.Context, imageURL string) error {
	if imageURL == "" {
		return nil // Nothing to delete
	}

	// Extract public ID from URL
	// Example URL: https://res.cloudinary.com/demo/image/upload/v1234567890/folder/public_id.jpg
	// We need to extract the "folder/public_id" part

	parts := strings.Split(imageURL, "/upload/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid Cloudinary URL format")
	}

	// Get the part after /upload/ and remove version and file extension
	publicIDWithVersion := parts[1]
	// Remove version if present (v1234567890/)
	publicIDParts := strings.SplitN(publicIDWithVersion, "/", 2)
	var publicID string
	if len(publicIDParts) > 1 && strings.HasPrefix(publicIDParts[0], "v") {
		publicID = publicIDParts[1]
	} else {
		publicID = publicIDWithVersion
	}

	// Remove file extension
	publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))

	log.Info().Str("public_id", publicID).Msg("Deleting image from Cloudinary")

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete from Cloudinary: %w", err)
	}

	log.Info().Str("public_id", publicID).Msg("Image deleted successfully")
	return nil
}
