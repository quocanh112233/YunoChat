package upload

import (
	"context"
	"fmt"

	"backend/internal/pkg/cloudinary"
)

var allowedImageMimes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

const maxAvatarSize = 10 * 1024 * 1024 // 10MB

func (uc *usecase) PresignAvatar(ctx context.Context, userID, mimeType string, fileSize int64) (*cloudinary.SignatureResponse, error) {
	if fileSize > maxAvatarSize {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	if !allowedImageMimes[mimeType] {
		return nil, fmt.Errorf("unsupported image format: %s", mimeType)
	}

	publicID := userID

	return uc.cloudinaryClient.GenerateSignature(publicID)
}
