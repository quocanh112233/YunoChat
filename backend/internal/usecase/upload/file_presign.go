package upload

import (
	"context"
	"fmt"
	"time"

	"backend/internal/pkg/r2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var allowedFileMimes = map[string]bool{
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
	"application/zip": true,
	"text/plain": true,
	"image/jpeg": true,
	"image/png": true,
	"image/webp": true,
	"image/gif": true,
}

var allowedVideoMimes = map[string]bool{
	"video/mp4": true,
	"video/quicktime": true,
}

const maxFileSize = 50 * 1024 * 1024 // 50MB
const maxVideoSize = 100 * 1024 * 1024 // 100MB

func (uc *usecase) PresignFile(ctx context.Context, userID, convIDStr, fileType, mimeType, fileName string, fileSize int64) (*r2.PresignedResponse, error) {
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// 1. Authorization: check if user is in conversation
	pgConvID := pgtype.UUID{Bytes: convID, Valid: true}
	pgUserUUID := pgtype.UUID{Bytes: userUUID, Valid: true}

	isMember, err := uc.convRepo.IsConversationMember(ctx, pgConvID, pgUserUUID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("forbidden: user is not a participant of this conversation")
	}

	// 2. Validation
	switch fileType {
	case "VIDEO":
		if fileSize > maxVideoSize {
			return nil, fmt.Errorf("video size exceeds 100MB limit")
		}
		if !allowedVideoMimes[mimeType] {
			return nil, fmt.Errorf("unsupported video format")
		}
	case "FILE", "IMAGE":
		if fileSize > maxFileSize {
			return nil, fmt.Errorf("file size exceeds 50MB limit")
		}
		if fileType == "FILE" && !allowedFileMimes[mimeType] {
			return nil, fmt.Errorf("unsupported file format")
		}
		if fileType == "IMAGE" && !allowedImageMimes[mimeType] {
			return nil, fmt.Errorf("unsupported image format")
		}
	default:
		return nil, fmt.Errorf("unsupported file type classification")
	}

	// 3. Generate R2 Key
	// Format: {type}/{conversation_id}/{YYYY/MM}/{uuid}_{original_name}
	now := time.Now()
	typeFolder := "files"
	if fileType == "VIDEO" {
		typeFolder = "videos"
	} else if fileType == "IMAGE" {
		typeFolder = "images"
	}

	fileUUID := uuid.New().String()
	r2Key := fmt.Sprintf("attachments/%s/%s/%04d/%02d/%s_%s", typeFolder, convIDStr, now.Year(), now.Month(), fileUUID, fileName)

	// 4. Generate Presigned URL
	return uc.r2Client.GeneratePresignedPutURL(ctx, r2Key, mimeType, fileSize)
}
