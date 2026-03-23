package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/repository/sqlc"
)

// Helper to parse UUID string to pgtype.UUID
func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// NotificationRepository defines notification operations
type NotificationRepository interface {
	Create(ctx context.Context, recipientID, actorID pgtype.UUID, notifType, referenceID, referenceType string) (sqlc.Notification, error)
	GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Notification, error)
	GetByReference(ctx context.Context, recipientID pgtype.UUID, notifType, referenceID string) (sqlc.Notification, error)
	ListByRecipient(ctx context.Context, recipientID pgtype.UUID, limit, offset int32) ([]sqlc.ListNotificationsByRecipientRow, error)
	MarkRead(ctx context.Context, id, recipientID pgtype.UUID) (sqlc.Notification, error)
	MarkAllRead(ctx context.Context, recipientID pgtype.UUID) error
	GetUnreadCount(ctx context.Context, recipientID pgtype.UUID) (int64, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteByReference(ctx context.Context, recipientID pgtype.UUID, notifType, referenceID string) error
}

// notificationRepository implements NotificationRepository using SQLC
type notificationRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(pool *pgxpool.Pool, queries *sqlc.Queries) NotificationRepository {
	return &notificationRepository{
		queries: queries,
		pool:    pool,
	}
}

func (r *notificationRepository) Create(ctx context.Context, recipientID, actorID pgtype.UUID, notifType, referenceID, referenceType string) (sqlc.Notification, error) {
	refID, err := parseUUID(referenceID)
	if err != nil {
		return sqlc.Notification{}, err
	}

	return r.queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		RecipientID:   recipientID,
		ActorID:       actorID,
		Type:          notifType,
		ReferenceID:   refID,
		ReferenceType: referenceType,
	})
}

func (r *notificationRepository) GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Notification, error) {
	return r.queries.GetNotificationByID(ctx, id)
}

func (r *notificationRepository) GetByReference(ctx context.Context, recipientID pgtype.UUID, notifType, referenceID string) (sqlc.Notification, error) {
	refID, err := parseUUID(referenceID)
	if err != nil {
		return sqlc.Notification{}, err
	}

	return r.queries.GetNotificationByReference(ctx, sqlc.GetNotificationByReferenceParams{
		RecipientID: recipientID,
		Type:        notifType,
		ReferenceID: refID,
	})
}

func (r *notificationRepository) ListByRecipient(ctx context.Context, recipientID pgtype.UUID, limit, offset int32) ([]sqlc.ListNotificationsByRecipientRow, error) {
	return r.queries.ListNotificationsByRecipient(ctx, sqlc.ListNotificationsByRecipientParams{
		RecipientID: recipientID,
		Limit:       limit,
		Offset:      offset,
	})
}

func (r *notificationRepository) MarkRead(ctx context.Context, id, recipientID pgtype.UUID) (sqlc.Notification, error) {
	return r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:          id,
		RecipientID: recipientID,
	})
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, recipientID pgtype.UUID) error {
	return r.queries.MarkAllNotificationsRead(ctx, recipientID)
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, recipientID pgtype.UUID) (int64, error) {
	return r.queries.GetUnreadNotificationCount(ctx, recipientID)
}

func (r *notificationRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteNotification(ctx, id)
}

func (r *notificationRepository) DeleteByReference(ctx context.Context, recipientID pgtype.UUID, notifType, referenceID string) error {
	refID, err := parseUUID(referenceID)
	if err != nil {
		return err
	}

	return r.queries.DeleteNotificationByReference(ctx, sqlc.DeleteNotificationByReferenceParams{
		RecipientID: recipientID,
		Type:        notifType,
		ReferenceID: refID,
	})
}
