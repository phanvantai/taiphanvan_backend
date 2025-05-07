package services

import (
	"context"
	"fmt"
	"mime/multipart"
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
	publicID := fmt.Sprintf("%s/user_%d_%d", s.cfg.UploadFolder, userID, timestamp)

	// Upload the file to Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "image",
		Folder:       s.cfg.UploadFolder,
	}

	log.Info().
		Str("public_id", publicID).
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
