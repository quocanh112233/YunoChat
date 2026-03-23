package message

import (
	"context"
	"errors"
	"time"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFriends         = errors.New("users are not friends - cannot send DM")
	ErrInvalidMessage     = errors.New("invalid message content")
	ErrNotMember          = errors.New("user is not a member of this conversation")
	ErrAttachmentRequired = errors.New("attachment data required for ATTACHMENT type")
)

// AttachmentData represents attachment information
type AttachmentData struct {
	StorageType  string
	FileType     string
	URL          string
	ThumbnailURL string
	OriginalName string
	MimeType     string
	SizeBytes    int64
	Width        *int32
	Height       *int32
	DurationSecs *int32
}

// SendMessageUseCase handles sending a new message
type SendMessageUseCase struct {
	convRepo postgres.ConversationRepository
	msgRepo  postgres.MessageRepository
	pool     *pgxpool.Pool
}

// NewSendMessageUseCase creates a new use case
func NewSendMessageUseCase(convRepo postgres.ConversationRepository, msgRepo postgres.MessageRepository, pool *pgxpool.Pool) *SendMessageUseCase {
	return &SendMessageUseCase{
		convRepo: convRepo,
		msgRepo:  msgRepo,
		pool:     pool,
	}
}

// SendMessageRequest represents the request parameters
type SendMessageRequest struct {
	SenderID       string
	ConversationID string
	Body           string
	Type           string // TEXT or ATTACHMENT
	Attachment     *AttachmentData
	ClientTempID   string // For optimistic UI
}

// MessageResponse represents a message in the response
type MessageResponse struct {
	ID             string          `json:"id"`
	ConversationID string          `json:"conversation_id"`
	Sender         SenderInfo      `json:"sender"`
	Body           string          `json:"body,omitempty"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	Attachment     *AttachmentInfo `json:"attachment,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
}

type SenderInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type AttachmentInfo struct {
	ID           string `json:"id"`
	StorageType  string `json:"storage_type"`
	FileType     string `json:"file_type"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Width        int32  `json:"width,omitempty"`
	Height       int32  `json:"height,omitempty"`
}

// SendMessageResponse represents the response
type SendMessageResponse struct {
	Message      MessageResponse `json:"message"`
	ClientTempID string          `json:"client_temp_id,omitempty"`
}

// Execute runs the use case
func (uc *SendMessageUseCase) Execute(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	// Validate message type and content
	if req.Type != "TEXT" && req.Type != "ATTACHMENT" {
		return nil, ErrInvalidMessage
	}

	if req.Type == "TEXT" && req.Body == "" {
		return nil, ErrInvalidMessage
	}

	if req.Type == "ATTACHMENT" && req.Attachment == nil {
		return nil, ErrAttachmentRequired
	}

	senderID, err := parseUUID(req.SenderID)
	if err != nil {
		return nil, err
	}

	conversationID, err := parseUUID(req.ConversationID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get conversation details to check type
	conv, err := uc.convRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// For DM, verify friendship
	if conv.Type == "DM" {
		// Get the other participant
		participants, err := uc.convRepo.GetConversationParticipants(ctx, conversationID)
		if err != nil {
			return nil, err
		}

		var otherUserID pgtype.UUID
		for _, p := range participants {
			if p.UserID != senderID {
				otherUserID = p.UserID
				break
			}
		}

		// Check friendship status
		status, err := uc.convRepo.GetFriendshipStatus(ctx, senderID, otherUserID)
		if err != nil || status != "ACCEPTED" {
			return nil, ErrNotFriends
		}
	}

	// Create message
	msgID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	msg, err := uc.msgRepo.CreateMessage(ctx, tx, msgID, conversationID, senderID, req.Body, req.Type, "SENT")
	if err != nil {
		return nil, err
	}

	// Create attachment if needed
	var attachment *AttachmentInfo
	if req.Type == "ATTACHMENT" && req.Attachment != nil {
		attID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		att, err := uc.msgRepo.CreateAttachment(ctx, tx, attID, msgID,
			req.Attachment.StorageType,
			req.Attachment.FileType,
			req.Attachment.URL,
			req.Attachment.ThumbnailURL,
			req.Attachment.OriginalName,
			req.Attachment.MimeType,
			req.Attachment.SizeBytes,
			req.Attachment.Width,
			req.Attachment.Height,
			req.Attachment.DurationSecs,
		)
		if err != nil {
			return nil, err
		}

		attachment = &AttachmentInfo{
			ID:           uuid.UUID(att.ID.Bytes).String(),
			StorageType:  att.StorageType,
			FileType:     att.FileType,
			URL:          att.Url,
			ThumbnailURL: att.ThumbnailUrl.String,
			OriginalName: att.OriginalName,
			MimeType:     att.MimeType,
			SizeBytes:    att.SizeBytes,
		}
		if att.Width.Valid {
			attachment.Width = att.Width.Int32
		}
		if att.Height.Valid {
			attachment.Height = att.Height.Int32
		}
	}

	// Update conversation last message and activity
	if err := uc.convRepo.UpdateLastMessage(ctx, tx, conversationID, msgID); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Send notification after commit (outside transaction)
	// Build notification payload
	payload := map[string]interface{}{
		"type":            "new_message",
		"conversation_id": req.ConversationID,
		"message_id":      uuid.UUID(msg.ID.Bytes).String(),
		"sender_id":       req.SenderID,
	}

	// For large payload handling, Hub will query DB if needed
	// Start a new transaction just for notification
	notifyTx, err := uc.pool.Begin(ctx)
	if err == nil {
		defer notifyTx.Rollback(ctx)
		_ = uc.msgRepo.NotifyNewMessage(ctx, notifyTx, "new_message", payload)
		_ = notifyTx.Commit(ctx)
	}

	// Build response
	response := &SendMessageResponse{
		Message: MessageResponse{
			ID:             uuid.UUID(msg.ID.Bytes).String(),
			ConversationID: req.ConversationID,
			Sender: SenderInfo{
				ID: req.SenderID,
			},
			Type:      msg.Type,
			Status:    msg.Status,
			CreatedAt: msg.CreatedAt.Time,
		},
		ClientTempID: req.ClientTempID,
	}

	if msg.Body.Valid {
		response.Message.Body = msg.Body.String
	}

	if attachment != nil {
		response.Message.Attachment = attachment
	}

	return response, nil
}

func parseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
