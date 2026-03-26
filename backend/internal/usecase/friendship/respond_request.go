package friendship

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"backend/internal/domain/user"
	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotAddressee   = errors.New("bạn không có quyền xử lý lời mời này")
	ErrAlreadyHandled = errors.New("lời mời này đã được xử lý")
	ErrInvalidAction  = errors.New("action không hợp lệ")
)

// RespondRequestUseCase handles accepting or declining friend requests
type RespondRequestUseCase struct {
	pool             *pgxpool.Pool
	friendshipRepo   postgres.FriendshipRepository
	conversationRepo postgres.ConversationRepository
	notificationRepo postgres.NotificationRepository
	userRepo         user.UserRepository
}

// NewRespondRequestUseCase creates a new RespondRequestUseCase
func NewRespondRequestUseCase(
	pool *pgxpool.Pool,
	friendshipRepo postgres.FriendshipRepository,
	conversationRepo postgres.ConversationRepository,
	notificationRepo postgres.NotificationRepository,
	userRepo user.UserRepository,
) *RespondRequestUseCase {
	return &RespondRequestUseCase{
		pool:             pool,
		friendshipRepo:   friendshipRepo,
		conversationRepo: conversationRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// RespondRequestRequest represents the request to respond to a friend request
type RespondRequestRequest struct {
	FriendshipID uuid.UUID
	UserID       uuid.UUID // The user responding (must be addressee)
	Action       string    // "ACCEPT" or "DECLINE"
}

// RespondAcceptResponse represents the response after accepting
type RespondAcceptResponse struct {
	FriendshipID   uuid.UUID `json:"friendship_id"`
	Status         string    `json:"status"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Friend         UserInfo  `json:"friend"`
}

// UserInfo represents simplified user info
type UserInfo struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
}

// Execute responds to a friend request (ACCEPT or DECLINE)
func (uc *RespondRequestUseCase) Execute(ctx context.Context, req RespondRequestRequest) (interface{}, error) {
	if req.Action != "ACCEPT" && req.Action != "DECLINE" {
		return nil, ErrInvalidAction
	}

	friendshipPgID := pgtype.UUID{Bytes: req.FriendshipID, Valid: true}
	userPgID := pgtype.UUID{Bytes: req.UserID, Valid: true}

	// Get friendship
	friendship, err := uc.friendshipRepo.GetByID(ctx, friendshipPgID)
	if err != nil {
		return nil, ErrRequestNotFound
	}

	// Verify user is the addressee
	if friendship.AddresseeID != userPgID {
		return nil, ErrNotAddressee
	}

	// Check if already handled
	if friendship.Status != "PENDING" {
		return nil, ErrAlreadyHandled
	}

	if req.Action == "DECLINE" {
		// Hard delete for decline
		err = uc.friendshipRepo.Delete(ctx, friendshipPgID)
		if err != nil {
			return nil, err
		}
		// Delete the notification
		_ = uc.notificationRepo.DeleteByReference(ctx, friendship.AddresseeID, "FRIEND_REQUEST", req.FriendshipID.String())

		return map[string]string{"message": "Đã từ chối lời mời kết bạn"}, nil
	}

	// ACCEPT: All in one transaction
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update friendship status
	_, err = tx.Exec(ctx,
		"UPDATE friendships SET status = 'ACCEPTED', updated_at = NOW() WHERE id = $1",
		req.FriendshipID,
	)
	if err != nil {
		return nil, err
	}

	// Find or create DM conversation
	requesterID := pgToUUID(friendship.RequesterID)

	var conversationID uuid.UUID
	existingConvID, err := uc.friendshipRepo.FindDMConversation(ctx, friendship.RequesterID, friendship.AddresseeID)
	if err == nil && existingConvID.Valid {
		// Reuse existing conversation
		conversationID = pgToUUID(existingConvID)
	} else {
		// Create new conversation
		convPgID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		_, err = tx.Exec(ctx,
			"INSERT INTO conversations (id, type, last_activity_at, created_at, updated_at) VALUES ($1, 'DM', NOW(), NOW(), NOW())",
			convPgID,
		)
		if err != nil {
			return nil, err
		}
		conversationID = pgToUUID(convPgID)

		// Add participants
		_, err = tx.Exec(ctx,
			"INSERT INTO conversation_participants (id, conversation_id, user_id, role, joined_at) VALUES (gen_random_uuid(), $1, $2, 'MEMBER', NOW())",
			convPgID, friendship.RequesterID,
		)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx,
			"INSERT INTO conversation_participants (id, conversation_id, user_id, role, joined_at) VALUES (gen_random_uuid(), $1, $2, 'MEMBER', NOW())",
			convPgID, friendship.AddresseeID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Create notification for requester (FRIEND_ACCEPTED)
	notifPgID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = tx.Exec(ctx,
		"INSERT INTO notifications (id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at) VALUES ($1, $2, $3, 'FRIEND_ACCEPTED', $4, 'friendship', FALSE, NOW())",
		notifPgID, friendship.RequesterID, friendship.AddresseeID, friendship.ID,
	)
	if err != nil {
		return nil, err
	}

	// Delete the FRIEND_REQUEST notification
	_, _ = tx.Exec(ctx,
		"DELETE FROM notifications WHERE recipient_id = $1 AND type = 'FRIEND_REQUEST' AND reference_id = $2",
		friendship.AddresseeID, friendship.ID,
	)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Broadcast via WS
	payload := map[string]interface{}{
		"type": "notification_new",
		"data": map[string]interface{}{
			"type":         "FRIEND_ACCEPTED",
			"reference_id": pgToUUID(friendship.ID).String(),
			"actor_id":     req.UserID.String(),
		},
		"recipient_ids": []string{pgToUUID(friendship.RequesterID).String()},
	}
	jsonPayload, _ := json.Marshal(payload)
	_, err = uc.pool.Exec(ctx, "SELECT pg_notify('chat_events', $1)", string(jsonPayload))
	if err != nil {
		log.Printf("Error sending friend accepted notification: %v", err)
	}

	// Get friend info
	friendUser, err := uc.userRepo.FindByID(ctx, requesterID)
	if err != nil {
		friendUser = &user.User{
			ID:          requesterID,
			Username:    "unknown",
			DisplayName: "Unknown User",
		}
	}

	return &RespondAcceptResponse{
		FriendshipID:   req.FriendshipID,
		Status:         "ACCEPTED",
		ConversationID: conversationID,
		Friend: UserInfo{
			ID:          friendUser.ID,
			Username:    friendUser.Username,
			DisplayName: friendUser.DisplayName,
			AvatarURL:   friendUser.AvatarURL,
		},
	}, nil
}

// Helper to convert pgtype.UUID to uuid.UUID
func pgToUUID(p pgtype.UUID) uuid.UUID {
	return p.Bytes
}
