package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Queries
	ListMessages(ctx context.Context, conversationID pgtype.UUID, cursorTime *pgtype.Timestamptz, cursorID *pgtype.UUID, limit int32) ([]sqlc.ListMessagesRow, error)
	GetMessageByID(ctx context.Context, id pgtype.UUID) (sqlc.GetMessageByIDRow, error)
	GetLatestMessageID(ctx context.Context, conversationID pgtype.UUID) (pgtype.UUID, error)
	GetUnreadCount(ctx context.Context, conversationID, userID pgtype.UUID) (int64, error)

	// Commands (with transaction support)
	CreateMessage(ctx context.Context, tx pgx.Tx, id, conversationID, senderID pgtype.UUID, body, msgType, status string) (sqlc.CreateMessageRow, error)
	CreateAttachment(ctx context.Context, tx pgx.Tx, id, messageID pgtype.UUID, storageType, fileType, url, thumbnailURL, originalName, mimeType string, sizeBytes int64, width, height, durationSecs *int32) (sqlc.Attachment, error)
	SoftDeleteMessage(ctx context.Context, messageID, senderID pgtype.UUID) (int64, error)
	UpdateMessageStatus(ctx context.Context, conversationID pgtype.UUID, status string, senderID pgtype.UUID) error

	// Notification
	NotifyNewMessage(ctx context.Context, tx pgx.Tx, eventType string, payload map[string]interface{}) error
}

// messageRepository implements MessageRepository
type messageRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(pool *pgxpool.Pool) MessageRepository {
	return &messageRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// ListMessages implements MessageRepository
func (r *messageRepository) ListMessages(ctx context.Context, conversationID pgtype.UUID, cursorTime *pgtype.Timestamptz, cursorID *pgtype.UUID, limit int32) ([]sqlc.ListMessagesRow, error) {
	params := sqlc.ListMessagesParams{
		ConversationID: conversationID,
		Limit:          limit,
	}

	if cursorTime != nil {
		params.Column2 = *cursorTime
	}
	if cursorID != nil {
		params.Column3 = *cursorID
	}

	return r.queries.ListMessages(ctx, params)
}

// GetMessageByID implements MessageRepository
func (r *messageRepository) GetMessageByID(ctx context.Context, id pgtype.UUID) (sqlc.GetMessageByIDRow, error) {
	return r.queries.GetMessageByID(ctx, id)
}

// GetLatestMessageID implements MessageRepository
func (r *messageRepository) GetLatestMessageID(ctx context.Context, conversationID pgtype.UUID) (pgtype.UUID, error) {
	return r.queries.GetLatestMessageID(ctx, conversationID)
}

// GetUnreadCount implements MessageRepository
func (r *messageRepository) GetUnreadCount(ctx context.Context, conversationID, userID pgtype.UUID) (int64, error) {
	return r.queries.GetUnreadCount(ctx, sqlc.GetUnreadCountParams{
		ConversationID: conversationID,
		SenderID:       userID,
	})
}

// CreateMessage implements MessageRepository
func (r *messageRepository) CreateMessage(ctx context.Context, tx pgx.Tx, id, conversationID, senderID pgtype.UUID, body, msgType, status string) (sqlc.CreateMessageRow, error) {
	queries := sqlc.New(tx)

	params := sqlc.CreateMessageParams{
		ID:             id,
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           msgType,
		Status:         status,
	}

	if body != "" {
		params.Body = pgtype.Text{String: body, Valid: true}
	}

	return queries.CreateMessage(ctx, params)
}

// CreateAttachment implements MessageRepository
func (r *messageRepository) CreateAttachment(ctx context.Context, tx pgx.Tx, id, messageID pgtype.UUID, storageType, fileType, url, thumbnailURL, originalName, mimeType string, sizeBytes int64, width, height, durationSecs *int32) (sqlc.Attachment, error) {
	queries := sqlc.New(tx)

	params := sqlc.CreateAttachmentParams{
		ID:           id,
		MessageID:    messageID,
		StorageType:  storageType,
		FileType:     fileType,
		Url:          url,
		OriginalName: originalName,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
	}

	if thumbnailURL != "" {
		params.ThumbnailUrl = pgtype.Text{String: thumbnailURL, Valid: true}
	}
	if width != nil {
		params.Width = pgtype.Int4{Int32: *width, Valid: true}
	}
	if height != nil {
		params.Height = pgtype.Int4{Int32: *height, Valid: true}
	}
	if durationSecs != nil {
		params.DurationSecs = pgtype.Int4{Int32: *durationSecs, Valid: true}
	}

	return queries.CreateAttachment(ctx, params)
}

// SoftDeleteMessage implements MessageRepository
func (r *messageRepository) SoftDeleteMessage(ctx context.Context, messageID, senderID pgtype.UUID) (int64, error) {
	return r.queries.SoftDeleteMessage(ctx, sqlc.SoftDeleteMessageParams{
		ID:       messageID,
		SenderID: senderID,
	})
}

// UpdateMessageStatus implements MessageRepository
func (r *messageRepository) UpdateMessageStatus(ctx context.Context, conversationID pgtype.UUID, status string, senderID pgtype.UUID) error {
	return r.queries.UpdateMessageStatus(ctx, sqlc.UpdateMessageStatusParams{
		ConversationID: conversationID,
		Status:         status,
		SenderID:       senderID,
	})
}

// NotifyNewMessage implements MessageRepository
// Sends pg_notify after transaction commit for real-time events
func (r *messageRepository) NotifyNewMessage(ctx context.Context, tx pgx.Tx, eventType string, payload map[string]interface{}) error {
	// Convert payload to JSON string
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notify payload: %w", err)
	}

	// Check payload size limit (8000 bytes for pg_notify, use 7500 as safety margin)
	if len(jsonPayload) > 7500 {
		// For large payloads, only send minimal info and let Hub query DB
		minimalPayload := fmt.Sprintf(`{"type":"%s","conversation_id":"%s","message_id":"%s"}`,
			eventType,
			payload["conversation_id"],
			payload["message_id"],
		)
		jsonPayload = []byte(minimalPayload)
	}

	_, err = tx.Exec(ctx, "SELECT pg_notify('chat_events', $1)", string(jsonPayload))
	return err
}
