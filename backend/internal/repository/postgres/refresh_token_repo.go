package postgres

import (
	"context"
	"errors"

	"backend/internal/domain/user"
	"backend/internal/repository/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type refreshTokenRepo struct {
	q *sqlc.Queries
}

func NewRefreshTokenRepository(q *sqlc.Queries) user.RefreshTokenRepository {
	return &refreshTokenRepo{q: q}
}

func (r *refreshTokenRepo) Create(ctx context.Context, t *user.RefreshToken) error {
	params := sqlc.CreateRefreshTokenParams{
		ID:        pgtype.UUID{Bytes: t.ID, Valid: true},
		UserID:    pgtype.UUID{Bytes: t.UserID, Valid: true},
		TokenHash: t.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: t.ExpiresAt, Valid: true},
	}
	dbToken, err := r.q.CreateRefreshToken(ctx, params)
	if err != nil {
		return err
	}

	t.IsRevoked = dbToken.IsRevoked
	t.CreatedAt = dbToken.CreatedAt.Time
	return nil
}

func (r *refreshTokenRepo) FindByHash(ctx context.Context, hash string) (*user.RefreshToken, error) {
	dbToken, err := r.q.FindRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("refresh token not found or revoked")
		}
		return nil, err
	}

	return &user.RefreshToken{
		ID:        dbToken.ID.Bytes,
		UserID:    dbToken.UserID.Bytes,
		TokenHash: dbToken.TokenHash,
		ExpiresAt: dbToken.ExpiresAt.Time,
		IsRevoked: dbToken.IsRevoked,
		CreatedAt: dbToken.CreatedAt.Time,
	}, nil
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.q.RevokeRefreshToken(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func (r *refreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.q.RevokeAllUserTokens(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return errors.New("method DeleteExpired not implemented")
}
