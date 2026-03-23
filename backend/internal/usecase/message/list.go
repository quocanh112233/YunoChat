package message

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"backend/internal/repository/postgres"
	"backend/internal/repository/sqlc"
)

var (
	ErrInvalidMessageCursor = errors.New("invalid cursor parameters")
)

// ListMessagesUseCase handles listing messages with cursor pagination
type ListMessagesUseCase struct {
	convRepo postgres.ConversationRepository
	msgRepo  postgres.MessageRepository
}

// NewListMessagesUseCase creates a new use case
func NewListMessagesUseCase(convRepo postgres.ConversationRepository, msgRepo postgres.MessageRepository) *ListMessagesUseCase {
	return &ListMessagesUseCase{
		convRepo: convRepo,
		msgRepo:  msgRepo,
	}
}

// ListMessagesRequest represents the request parameters
type ListMessagesRequest struct {
	UserID         string
	ConversationID string
	Limit          int32
	BeforeID       *string
	BeforeTime     *string
}

// ListMessagesResponse represents the response
type ListMessagesResponse struct {
	Messages []MessageResponse      `json:"messages"`
	Meta     CursorPaginationMeta   `json:"meta"`
}

// CursorPaginationMeta for message pagination
type CursorPaginationMeta struct {
	HasMore    bool       `json:"has_more"`
	NextCursor *NextCursor `json:"next_cursor,omitempty"`
}

type NextCursor struct {
	BeforeID   string `json:"before_id"`
	BeforeTime string `json:"before_time"`
}

// Execute runs the use case
func (uc *ListMessagesUseCase) Execute(ctx context.Context, req ListMessagesRequest) (*ListMessagesResponse, error) {
	// Validate and set default limit
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 30
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		return nil, err
	}

	conversationID, err := parseUUID(req.ConversationID)
	if err != nil {
		return nil, err
	}

	// Verify user is a member
	isMember, err := uc.convRepo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	// Parse cursor parameters
	var cursorTime *pgtype.Timestamptz
	var cursorID *pgtype.UUID

	if req.BeforeTime != nil && req.BeforeID != nil {
		t, err := time.Parse(time.RFC3339, *req.BeforeTime)
		if err != nil {
			return nil, ErrInvalidMessageCursor
		}
		cursorTime = &pgtype.Timestamptz{Time: t, Valid: true}

		id, err := uuid.Parse(*req.BeforeID)
		if err != nil {
			return nil, ErrInvalidMessageCursor
		}
		cursorID = &pgtype.UUID{Bytes: id, Valid: true}
	}

	// Fetch messages from repository (DESC order - newest first)
	rows, err := uc.msgRepo.ListMessages(ctx, conversationID, cursorTime, cursorID, limit+1)
	if err != nil {
		return nil, err
	}

	// Check if there are more results
	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	// Convert to response format and reverse to ASC order (oldest first for chat UI)
	messages := make([]MessageResponse, len(rows))
	for i, row := range rows {
		msg := uc.convertToMessageResponse(row)
		// Reverse order: place at the end
		messages[len(rows)-1-i] = msg
	}

	// Build pagination meta
	meta := CursorPaginationMeta{HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		// Get the oldest message in the batch for next cursor
		oldestRow := rows[len(rows)-1]
		meta.NextCursor = &NextCursor{
			BeforeID:   uuid.UUID(oldestRow.ID.Bytes).String(),
			BeforeTime: oldestRow.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return &ListMessagesResponse{
		Messages: messages,
		Meta:     meta,
	}, nil
}

func (uc *ListMessagesUseCase) convertToMessageResponse(row sqlc.ListMessagesRow) MessageResponse {
	msg := MessageResponse{
		ID:             uuid.UUID(row.ID.Bytes).String(),
		ConversationID: uuid.UUID(row.ConversationID.Bytes).String(),
		Sender: SenderInfo{
			ID:          uuid.UUID(row.SenderID.Bytes).String(),
			DisplayName: row.SenderDisplayName,
			Username:    row.SenderUsername,
		},
		Type:      row.Type,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
	}

	if row.SenderAvatarUrl.Valid {
		msg.Sender.AvatarURL = row.SenderAvatarUrl.String
	}

	if row.Body.Valid && row.DeletedAt.Time.IsZero() {
		msg.Body = row.Body.String
	}

	// Add attachment info if exists
	if row.AttachmentID.Valid {
		msg.Attachment = &AttachmentInfo{
			ID:           uuid.UUID(row.AttachmentID.Bytes).String(),
			StorageType:  row.StorageType.String,
			FileType:     row.FileType.String,
			URL:          row.Url.String,
			ThumbnailURL: row.ThumbnailUrl.String,
			OriginalName: row.OriginalName.String,
			MimeType:     row.MimeType.String,
			SizeBytes:    row.SizeBytes.Int64,
		}
		if row.Width.Valid {
			msg.Attachment.Width = row.Width.Int32
		}
		if row.Height.Valid {
			msg.Attachment.Height = row.Height.Int32
		}
	}

	return msg
}
