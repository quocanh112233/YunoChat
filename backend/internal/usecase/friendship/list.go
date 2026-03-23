package friendship

import (
	"context"
	"time"

	"backend/internal/domain/user"
	"backend/internal/repository/postgres"
	"backend/internal/repository/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ListUseCase handles listing friends and pending requests
type ListUseCase struct {
	friendshipRepo postgres.FriendshipRepository
	userRepo       user.UserRepository
}

// NewListUseCase creates a new ListUseCase
func NewListUseCase(
	friendshipRepo postgres.FriendshipRepository,
	userRepo user.UserRepository,
) *ListUseCase {
	return &ListUseCase{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

// FriendInfo represents a friend in the list
type FriendInfo struct {
	FriendshipID    uuid.UUID  `json:"friendship_id"`
	User            UserInfo   `json:"user"`
	ConversationID  *uuid.UUID `json:"conversation_id,omitempty"`
	BecameFriendsAt time.Time  `json:"became_friends_at"`
}

// PendingRequestInfo represents a pending friend request
type PendingRequestInfo struct {
	RequestID   uuid.UUID `json:"request_id"`
	FromUser    UserInfo  `json:"from_user"`
	RequestedAt time.Time `json:"requested_at"`
}

// SentRequestInfo represents a sent friend request
type SentRequestInfo struct {
	RequestID   uuid.UUID `json:"request_id"`
	ToUser      UserInfo  `json:"to_user"`
	RequestedAt time.Time `json:"requested_at"`
}

// ListFriends lists all accepted friends
func (uc *ListUseCase) ListFriends(ctx context.Context, userID uuid.UUID) ([]FriendInfo, error) {
	userPgID := pgtype.UUID{Bytes: userID, Valid: true}

	rows, err := uc.friendshipRepo.ListFriends(ctx, userPgID)
	if err != nil {
		return nil, err
	}

	friends := make([]FriendInfo, 0, len(rows))
	for _, row := range rows {
		// Type assertion for FriendID which is interface{}
		var friendID uuid.UUID
		if friendIDBytes, ok := row.FriendID.([16]byte); ok {
			friendID = friendIDBytes
		} else if friendIDPg, ok := row.FriendID.(pgtype.UUID); ok {
			friendID = friendIDPg.Bytes
		}

		// Find DM conversation
		friendPgID := pgtype.UUID{Bytes: friendID, Valid: true}
		convID, _ := uc.friendshipRepo.FindDMConversation(ctx, userPgID, friendPgID)

		var convUUID *uuid.UUID
		if convID.Valid {
			id := pgToUUID(convID)
			convUUID = &id
		}

		var avatarURL *string
		if row.AvatarUrl.Valid {
			avatarURL = &row.AvatarUrl.String
		}

		friends = append(friends, FriendInfo{
			FriendshipID: pgToUUID(row.FriendshipID),
			User: UserInfo{
				ID:          friendID,
				Username:    row.Username,
				DisplayName: row.DisplayName,
				AvatarURL:   avatarURL,
			},
			ConversationID:  convUUID,
			BecameFriendsAt: row.BecameFriendsAt.Time,
		})
	}

	return friends, nil
}

// ListPendingReceived lists pending friend requests received
func (uc *ListUseCase) ListPendingReceived(ctx context.Context, userID uuid.UUID) ([]PendingRequestInfo, error) {
	userPgID := pgtype.UUID{Bytes: userID, Valid: true}

	rows, err := uc.friendshipRepo.ListPendingReceived(ctx, userPgID)
	if err != nil {
		return nil, err
	}

	requests := make([]PendingRequestInfo, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.AvatarUrl.Valid {
			avatarURL = &row.AvatarUrl.String
		}

		requests = append(requests, PendingRequestInfo{
			RequestID: pgToUUID(row.RequestID),
			FromUser: UserInfo{
				ID:          pgToUUID(row.FromUserID),
				Username:    row.Username,
				DisplayName: row.DisplayName,
				AvatarURL:   avatarURL,
			},
			RequestedAt: row.RequestedAt.Time,
		})
	}

	return requests, nil
}

// ListPendingSent lists pending friend requests sent
func (uc *ListUseCase) ListPendingSent(ctx context.Context, userID uuid.UUID) ([]SentRequestInfo, error) {
	userPgID := pgtype.UUID{Bytes: userID, Valid: true}

	rows, err := uc.friendshipRepo.ListPendingSent(ctx, userPgID)
	if err != nil {
		return nil, err
	}

	requests := make([]SentRequestInfo, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.AvatarUrl.Valid {
			avatarURL = &row.AvatarUrl.String
		}

		requests = append(requests, SentRequestInfo{
			RequestID: pgToUUID(row.RequestID),
			ToUser: UserInfo{
				ID:          pgToUUID(row.ToUserID),
				Username:    row.Username,
				DisplayName: row.DisplayName,
				AvatarURL:   avatarURL,
			},
			RequestedAt: row.RequestedAt.Time,
		})
	}

	return requests, nil
}

// SearchUsers searches users with relationship status
func (uc *ListUseCase) SearchUsers(ctx context.Context, currentUserID uuid.UUID, query string, limit int32) ([]sqlc.SearchUsersWithRelationshipRow, error) {
	userPgID := pgtype.UUID{Bytes: currentUserID, Valid: true}
	return uc.friendshipRepo.SearchUsersWithRelationship(ctx, userPgID, query, limit)
}
