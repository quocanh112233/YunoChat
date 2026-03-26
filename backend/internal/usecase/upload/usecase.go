package upload

import (
	"context"

	"backend/internal/pkg/cloudinary"
	"backend/internal/pkg/r2"
	"backend/internal/repository/postgres"
)

type Usecase interface {
	PresignAvatar(ctx context.Context, userID, mimeType string, fileSize int64) (*cloudinary.SignatureResponse, error)
	PresignFile(ctx context.Context, userID, convID, fileType, mimeType, fileName string, fileSize int64) (*r2.PresignedResponse, error)
}

type usecase struct {
	cloudinaryClient cloudinary.Client
	r2Client         r2.Client
	convRepo         postgres.ConversationRepository
}

func NewUsecase(cldClient cloudinary.Client, r2Client r2.Client, convRepo postgres.ConversationRepository) Usecase {
	return &usecase{
		cloudinaryClient: cldClient,
		r2Client:         r2Client,
		convRepo:         convRepo,
	}
}
