package notification

import (
	"context"
	"time"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ListUseCase handles listing notifications
type ListUseCase struct {
	notificationRepo postgres.NotificationRepository
}

// NewListUseCase creates a new ListUseCase
func NewListUseCase(notificationRepo postgres.NotificationRepository) *ListUseCase {
	return &ListUseCase{
		notificationRepo: notificationRepo,
	}
}

// ActorInfo represents notification actor info
type ActorInfo struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
}

// NotificationInfo represents a notification
type NotificationInfo struct {
	ID            uuid.UUID  `json:"id"`
	Type          string     `json:"type"`
	IsRead        bool       `json:"is_read"`
	Actor         ActorInfo  `json:"actor"`
	ReferenceID   uuid.UUID  `json:"reference_id"`
	ReferenceType string     `json:"reference_type"`
	PreviewText   string     `json:"preview_text"`
	CreatedAt     time.Time  `json:"created_at"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
}

// ListResponse represents the response for listing notifications
type ListResponse struct {
	Notifications []NotificationInfo `json:"data"`
	Meta          ListMeta           `json:"meta"`
}

// ListMeta represents pagination metadata
type ListMeta struct {
	UnreadCount int64 `json:"unread_count"`
	Total       int   `json:"total"`
	HasMore     bool  `json:"has_more"`
}

// ListRequest represents the request for listing notifications
type ListRequest struct {
	UserID uuid.UUID
	Limit  int32
	Offset int32
}

// Execute lists notifications for a user
func (uc *ListUseCase) Execute(ctx context.Context, req ListRequest) (*ListResponse, error) {
	userPgID := pgtype.UUID{Bytes: req.UserID, Valid: true}

	// Get notifications
	rows, err := uc.notificationRepo.ListByRecipient(ctx, userPgID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// Get unread count
	unreadCount, err := uc.notificationRepo.GetUnreadCount(ctx, userPgID)
	if err != nil {
		unreadCount = 0
	}

	notifications := make([]NotificationInfo, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.ActorAvatarUrl.Valid {
			avatarURL = &row.ActorAvatarUrl.String
		}

		var readAt *time.Time
		if row.ReadAt.Valid {
			readAt = &row.ReadAt.Time
		}

		// Generate preview text based on type
		previewText := generatePreviewText(row.Type, row.ActorDisplayName)

		notifications = append(notifications, NotificationInfo{
			ID:     pgToUUID(row.ID),
			Type:   row.Type,
			IsRead: row.IsRead,
			Actor: ActorInfo{
				ID:          pgToUUID(row.ActorID),
				Username:    row.ActorUsername,
				DisplayName: row.ActorDisplayName,
				AvatarURL:   avatarURL,
			},
			ReferenceID:   pgToUUID(row.ReferenceID),
			ReferenceType: row.ReferenceType,
			PreviewText:   previewText,
			CreatedAt:     row.CreatedAt.Time,
			ReadAt:        readAt,
		})
	}

	hasMore := len(rows) == int(req.Limit)

	return &ListResponse{
		Notifications: notifications,
		Meta: ListMeta{
			UnreadCount: unreadCount,
			Total:       len(notifications),
			HasMore:     hasMore,
		},
	}, nil
}

// generatePreviewText generates preview text based on notification type
func generatePreviewText(notifType, actorName string) string {
	switch notifType {
	case "FRIEND_REQUEST":
		return actorName + " đã gửi lời mời kết bạn"
	case "FRIEND_ACCEPTED":
		return actorName + " đã chấp nhận lời mời kết bạn"
	case "GROUP_ADDED":
		return actorName + " đã thêm bạn vào nhóm"
	default:
		return "Bạn có thông báo mới"
	}
}

// pgToUUID converts pgtype.UUID to uuid.UUID
func pgToUUID(p pgtype.UUID) uuid.UUID {
	return p.Bytes
}
