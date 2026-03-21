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

type RegisterInput struct {
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username" validate:"required,min=3,max=30"`
	Password    string `json:"password" validate:"required,min=8,max=255"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

type RegisterOutput struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type RegisterUseCase struct {
	userRepo  user.UserRepository
	tokenRepo user.RefreshTokenRepository
	cfg       *config.Config
	val       *validator.CustomValidator
}

func NewRegisterUseCase(
	ur user.UserRepository,
	tr user.RefreshTokenRepository,
	cfg *config.Config,
	val *validator.CustomValidator,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:  ur,
		tokenRepo: tr,
		cfg:       cfg,
		val:       val,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error, map[string]string) {
	// Validate input
	if errMap := uc.val.ValidateStruct(input); errMap != nil {
		return nil, nil, errMap
	}

	// Check duplicates
	if existing, _ := uc.userRepo.FindByEmail(ctx, input.Email); existing != nil {
		return nil, user.ErrDuplicateEmail, nil
	}
	if existing, _ := uc.userRepo.FindByUsername(ctx, input.Username); existing != nil {
		return nil, user.ErrDuplicateUsername, nil
	}

	// Hash password
	hashedPwd, err := password.Hash(input.Password)
	if err != nil {
		return nil, err, nil
	}

	// Create user
	u := &user.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: hashedPwd,
		DisplayName:  input.DisplayName,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, err, nil
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

	return &RegisterOutput{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int(accTTL.Seconds()),
	}, nil, nil
}
