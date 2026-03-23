package conversation

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
	ErrNotMember     = errors.New("user is not a member of this conversation")
	ErrNotAdmin      = errors.New("user is not an admin of this group")
	ErrInvalidCursor = errors.New("invalid cursor parameters")
)

// ConversationWithDetails represents a conversation with full details for API response
type ConversationWithDetails struct {
	ID             string                `json:"id"`
	Type           string                `json:"type"`
	Name           string                `json:"name,omitempty"`
	AvatarURL      string                `json:"avatar_url,omitempty"`
	LastActivityAt time.Time             `json:"last_activity_at"`
	LastMessage    *LastMessageInfo      `json:"last_message,omitempty"`
	UnreadCount    int64                 `json:"unread_count"`
	OtherUser      *OtherUserInfo        `json:"other_user,omitempty"` // For DM only
	Participants   []ParticipantInfo       `json:"participants,omitempty"`
}

type LastMessageInfo struct {
	ID        string    `json:"id"`
	Body      string    `json:"body,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Sender    SenderInfo `json:"sender"`
}

type SenderInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type OtherUserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Status      string `json:"status"`
}

type ParticipantInfo struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Status      string    `json:"status"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// CursorPaginationMeta for API response
type CursorPaginationMeta struct {
	HasMore    bool       `json:"has_more"`
	NextCursor *NextCursor `json:"next_cursor,omitempty"`
}

type NextCursor struct {
	BeforeID   string `json:"before_id"`
	BeforeTime string `json:"before_time"`
}

// ListConversationsUseCase handles listing conversations with cursor pagination
type ListConversationsUseCase struct {
	convRepo postgres.ConversationRepository
}

// NewListConversationsUseCase creates a new use case
func NewListConversationsUseCase(convRepo postgres.ConversationRepository) *ListConversationsUseCase {
	return &ListConversationsUseCase{convRepo: convRepo}
}

// ListConversationsRequest represents the request parameters
type ListConversationsRequest struct {
	UserID     string
	Limit      int32
	BeforeID   *string
	BeforeTime *string
}

// ListConversationsResponse represents the response
type ListConversationsResponse struct {
	Conversations []ConversationWithDetails
	Meta          CursorPaginationMeta
}

// Execute runs the use case
func (uc *ListConversationsUseCase) Execute(ctx context.Context, req ListConversationsRequest) (*ListConversationsResponse, error) {
	// Validate and set default limit
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 30
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		return nil, err
	}

	// Parse cursor parameters
	var cursorTime *pgtype.Timestamptz
	var cursorID *pgtype.UUID

	if req.BeforeTime != nil && req.BeforeID != nil {
		t, err := time.Parse(time.RFC3339, *req.BeforeTime)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		cursorTime = &pgtype.Timestamptz{Time: t, Valid: true}

		id, err := uuid.Parse(*req.BeforeID)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		cursorID = &pgtype.UUID{Bytes: id, Valid: true}
	}

	// Fetch conversations from repository
	rows, err := uc.convRepo.ListConversationsByUser(ctx, userID, cursorTime, cursorID, limit+1)
	if err != nil {
		return nil, err
	}

	// Check if there are more results
	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit] // Remove the extra item
	}

	// Convert to response format
	conversations := make([]ConversationWithDetails, len(rows))
	for i, row := range rows {
		conv := uc.convertToConversationWithDetails(row)
		conversations[i] = conv
	}

	// Build pagination meta
	meta := CursorPaginationMeta{HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		lastRow := rows[len(rows)-1]
		meta.NextCursor = &NextCursor{
			BeforeID:   uuid.UUID(lastRow.ID.Bytes).String(),
			BeforeTime: lastRow.LastActivityAt.Time.Format(time.RFC3339),
		}
	}

	return &ListConversationsResponse{
		Conversations: conversations,
		Meta:          meta,
	}, nil
}

func (uc *ListConversationsUseCase) convertToConversationWithDetails(row sqlc.ListConversationsByUserRow) ConversationWithDetails {
	conv := ConversationWithDetails{
		ID:             uuid.UUID(row.ID.Bytes).String(),
		Type:           row.Type,
		LastActivityAt: row.LastActivityAt.Time,
		UnreadCount:    row.UnreadCount,
	}

	if row.Name.Valid {
		conv.Name = row.Name.String
	}
	if row.AvatarUrl.Valid {
		conv.AvatarURL = row.AvatarUrl.String
	}

	// Add last message info if exists
	if row.LastMessageID.Valid {
		conv.LastMessage = &LastMessageInfo{
			ID:        uuid.UUID(row.LastMessageID.Bytes).String(),
			Type:      row.LastMessageType.String,
			CreatedAt: row.LastMessageCreatedAt.Time,
			Sender: SenderInfo{
				ID:          uuid.UUID(row.LastMessageSenderID.Bytes).String(),
				DisplayName: row.LastMessageSenderName.String,
			},
		}
		if row.LastMessageBody.Valid {
			conv.LastMessage.Body = row.LastMessageBody.String
		}
		if row.LastMessageSenderAvatar.Valid {
			conv.LastMessage.Sender.AvatarURL = row.LastMessageSenderAvatar.String
		}
	}

	return conv
}

func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
