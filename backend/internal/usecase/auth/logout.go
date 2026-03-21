package auth

import (
	"context"

	"backend/internal/domain/user"

	"github.com/google/uuid"
)

type LogoutInput struct {
	UserID       uuid.UUID
	RefreshToken string
}

type LogoutUseCase struct {
	userRepo  user.UserRepository
	tokenRepo user.RefreshTokenRepository
}

func NewLogoutUseCase(
	ur user.UserRepository,
	tr user.RefreshTokenRepository,
) *LogoutUseCase {
	return &LogoutUseCase{
		userRepo:  ur,
		tokenRepo: tr,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	// Revoke the specific token
	incomingHash := HashToken(input.RefreshToken)
	dbToken, err := uc.tokenRepo.FindByHash(ctx, incomingHash)
	if err == nil && dbToken != nil {
		_ = uc.tokenRepo.Revoke(ctx, dbToken.ID)
	}

	// Update presence to OFFLINE
	_ = uc.userRepo.UpdatePresence(ctx, input.UserID, "OFFLINE")

	return nil
}
