package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"backend/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCannotSelf = errors.New("cannot kick yourself")
)

// KickMemberUseCase handles kicking a member from a group conversation
type KickMemberUseCase struct {
	convRepo postgres.ConversationRepository
	pool     *pgxpool.Pool
}

// NewKickMemberUseCase creates a new use case
func NewKickMemberUseCase(convRepo postgres.ConversationRepository, pool *pgxpool.Pool) *KickMemberUseCase {
	return &KickMemberUseCase{
		convRepo: convRepo,
		pool:     pool,
	}
}

// KickMemberRequest represents the request parameters
type KickMemberRequest struct {
	CallerID       string
	ConversationID string
	TargetUserID   string
}

// Execute runs the use case
func (uc *KickMemberUseCase) Execute(ctx context.Context, req KickMemberRequest) error {
	if req.CallerID == req.TargetUserID {
		return ErrCannotSelf
	}

	callerID, err := parseUUID(req.CallerID)
	if err != nil {
		return err
	}
	convID, err := parseUUID(req.ConversationID)
	if err != nil {
		return err
	}
	targetID, err := parseUUID(req.TargetUserID)
	if err != nil {
		return err
	}

	// Check caller is admin
	isAdmin, err := uc.convRepo.IsGroupAdmin(ctx, convID, callerID)
	if err != nil || !isAdmin {
		return ErrNotAdmin
	}

	// Check target is a member
	isMember, err := uc.convRepo.IsConversationMember(ctx, convID, targetID)
	if err != nil || !isMember {
		return ErrNotMember
	}

	// Remove member: set left_at = NOW()
	_, err = uc.pool.Exec(ctx,
		`UPDATE conversation_participants
		 SET left_at = NOW()
		 WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL`,
		pgtype.UUID{Bytes: [16]byte(convID.Bytes), Valid: true},
		pgtype.UUID{Bytes: [16]byte(targetID.Bytes), Valid: true},
	)
	if err != nil {
		return err
	}

	// Broadcast WS event: member_removed
	// In a real app, this should be done via pg_notify to reach all Hub instances
	participants, _ := uc.convRepo.GetConversationParticipants(ctx, convID)
	recipientIDs := make([]string, 0, len(participants)+1)
	recipientIDs = append(recipientIDs, req.TargetUserID) // Include kicked user
	for _, p := range participants {
		recipientIDs = append(recipientIDs, pgtypeUUIDToString(p.UserID))
	}

	// Get caller info for AddedBy
	caller, _ := uc.convRepo.GetConversationParticipants(ctx, convID) // Reuse GetParticipants to find caller
	var callerName string
	for _, p := range caller {
		if pgtypeUUIDToString(p.UserID) == req.CallerID {
			callerName = p.DisplayName
			break
		}
	}

	event := map[string]interface{}{
		"type": "member_removed",
		"data": map[string]interface{}{
			"conversation_id": req.ConversationID,
			"removed_user_id": req.TargetUserID,
			"removed_by": map[string]string{
				"id":           req.CallerID,
				"display_name": callerName,
			},
			"reason": "KICKED",
		},
		"recipient_ids": recipientIDs,
	}

	jsonPayload, _ := json.Marshal(event)
	_, err = uc.pool.Exec(ctx, "SELECT pg_notify('chat_events', $1)", string(jsonPayload))
	if err != nil {
		log.Printf("Error sending member_removed notification: %v", err)
	}

	return nil
}

func pgtypeUUIDToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	u, _ := uuid.FromBytes(id.Bytes[:])
	return u.String()
}
