package postgres

import (
	"context"
	"errors"
	"time"

	"backend/internal/domain/user"
	"backend/internal/repository/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type userRepo struct {
	q *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) user.UserRepository {
	return &userRepo{q: q}
}

// Helpers
func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgToUuid(p pgtype.UUID) uuid.UUID {
	return p.Bytes
}

func textToStrPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func tsToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func (r *userRepo) Create(ctx context.Context, u *user.User) error {
	params := sqlc.CreateUserParams{
		ID:           uuidToPg(u.ID),
		Email:        u.Email,
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
	}
	dbUser, err := r.q.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	u.Status = dbUser.Status
	u.CreatedAt = dbUser.CreatedAt.Time
	u.UpdatedAt = dbUser.UpdatedAt.Time
	u.LastSeenAt = tsToTime(dbUser.LastSeenAt)
	return nil
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	dbUser, err := r.q.FindUserByID(ctx, uuidToPg(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapToDomainUser(dbUser), nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	dbUser, err := r.q.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapToDomainUser(dbUser), nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	dbUser, err := r.q.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapToDomainUser(dbUser), nil
}

func (r *userRepo) Update(ctx context.Context, u *user.User) error {
	return errors.New("method Update not implemented")
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return errors.New("method Delete not implemented")
}

func (r *userRepo) UpdatePresence(ctx context.Context, id uuid.UUID, status string) error {
	params := sqlc.UpdateUserStatusParams{
		Status: status,
		ID:     uuidToPg(id),
	}
	return r.q.UpdateUserStatus(ctx, params)
}

func (r *userRepo) Search(ctx context.Context, query string, limit, offset int) ([]*user.User, error) {
	return nil, errors.New("method Search not implemented")
}

func mapToDomainUser(dbUser sqlc.User) *user.User {
	return &user.User{
		ID:                 pgToUuid(dbUser.ID),
		Email:              dbUser.Email,
		Username:           dbUser.Username,
		PasswordHash:       dbUser.PasswordHash,
		DisplayName:        dbUser.DisplayName,
		AvatarURL:          textToStrPtr(dbUser.AvatarUrl),
		Bio:                textToStrPtr(dbUser.Bio),
		AvatarCloudinaryID: textToStrPtr(dbUser.AvatarCloudinaryID),
		Status:             dbUser.Status,
		LastSeenAt:         tsToTime(dbUser.LastSeenAt),
		CreatedAt:          dbUser.CreatedAt.Time,
		UpdatedAt:          dbUser.UpdatedAt.Time,
	}
}
