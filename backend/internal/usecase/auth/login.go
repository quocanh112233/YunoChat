package auth

import (
	"context"
	"time"

	"backend/internal/config"
	"backend/internal/domain/user"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"backend/internal/pkg/validator"

	"github.com/google/uuid"
)

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginOutput struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type LoginUseCase struct {
	userRepo  user.UserRepository
	tokenRepo user.RefreshTokenRepository
	cfg       *config.Config
	val       *validator.CustomValidator
}

func NewLoginUseCase(
	ur user.UserRepository,
	tr user.RefreshTokenRepository,
	cfg *config.Config,
	val *validator.CustomValidator,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:  ur,
		tokenRepo: tr,
		cfg:       cfg,
		val:       val,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error, map[string]string) {
	// Validate input
	if errMap := uc.val.ValidateStruct(input); errMap != nil {
		return nil, nil, errMap
	}

	// Find User
	u, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Prevent timing attacks
		password.PreventTimingAttack(input.Password)
		return nil, user.ErrInvalidCredentials, nil
	}

	// Check password
	if err := password.Compare(u.PasswordHash, input.Password); err != nil {
		return nil, user.ErrInvalidCredentials, nil
	}

	// Generate Tokens
	accTTL := 15 * time.Minute
	accessToken, err := jwt.GenerateAccessToken(u.ID, uc.cfg.JWT.AccessSecret, accTTL)
	if err != nil {
		return nil, err, nil
	}

	refTTL := 7 * 24 * time.Hour
	refreshTokenRaw, err := jwt.GenerateAccessToken(u.ID, uc.cfg.JWT.RefreshSecret, refTTL)
	if err != nil {
		return nil, err, nil
	}

	refHash := HashToken(refreshTokenRaw)

	// Save Refresh Token
	rt := &user.RefreshToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: refHash,
		ExpiresAt: time.Now().Add(refTTL),
	}
	if err := uc.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err, nil
	}

	// Update presence status
	_ = uc.userRepo.UpdatePresence(ctx, u.ID, "ONLINE")

	return &LoginOutput{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int(accTTL.Seconds()),
	}, nil, nil
}
