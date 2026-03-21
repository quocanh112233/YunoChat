package auth

import (
	"context"
	"errors"
	"time"

	"backend/internal/config"
	"backend/internal/domain/user"
	"backend/internal/pkg/jwt"

	"github.com/google/uuid"
)

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type RefreshTokenUseCase struct {
	tokenRepo user.RefreshTokenRepository
	cfg       *config.Config
}

func NewRefreshTokenUseCase(
	tr user.RefreshTokenRepository,
	cfg *config.Config,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tokenRepo: tr,
		cfg:       cfg,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, cookieToken string) (*RefreshOutput, error) {
	if cookieToken == "" {
		return nil, errors.New("missing refresh token")
	}

	// 1. Verify that the cookie token is a valid JWT
	userID, err := jwt.ParseToken(cookieToken, uc.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// 2. Hash it to find it in the DB
	incomingHash := HashToken(cookieToken)
	dbToken, err := uc.tokenRepo.FindByHash(ctx, incomingHash)
	
	if err != nil || dbToken == nil {
		// Could not find valid token (either missing, revoked or expired)
		// Assuming anti-reuse: token was potentially revoked. Revoke all.
		_ = uc.tokenRepo.RevokeAllForUser(ctx, userID)
		return nil, errors.New("refresh token reused or invalid, all sessions revoked")
	}

	// 3. Token is valid. Proceed to ROTATE.
	// Revoke the old token.
	_ = uc.tokenRepo.Revoke(ctx, dbToken.ID)

	// 4. Generate Tokens
	accTTL := 15 * time.Minute
	accessToken, err := jwt.GenerateAccessToken(userID, uc.cfg.JWT.AccessSecret, accTTL)
	if err != nil {
		return nil, err
	}

	refTTL := 7 * 24 * time.Hour
	refreshTokenRaw, err := jwt.GenerateAccessToken(userID, uc.cfg.JWT.RefreshSecret, refTTL)
	if err != nil {
		return nil, err
	}

	// 5. Hash and Save new refresh token
	newRefHash := HashToken(refreshTokenRaw)

	rt := &user.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: newRefHash,
		ExpiresAt: time.Now().Add(refTTL),
	}
	if err := uc.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int(accTTL.Seconds()),
	}, nil
}
