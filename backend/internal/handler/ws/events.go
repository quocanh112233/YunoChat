package ws

// Event type constants for WebSocket communication
const (
	// Client -> Server events
	EventPing              = "ping"
	EventJoinConversation  = "join_conversation"
	EventLeaveConversation = "leave_conversation"
	EventSendMessage       = "send_message"
	EventTyping            = "typing"
	EventMarkRead          = "mark_read"

	// Server -> Client events
	EventConnected           = "connected"
	EventPong                = "pong"
	EventNewMessage          = "new_message"
	EventMessageSent         = "message_sent"
	EventUserTyping          = "user_typing"
	EventReadReceipt         = "read_receipt"
	EventPresenceUpdate      = "presence_update"
	EventNotificationNew     = "notification_new"
	EventConversationUpdated = "conversation_updated"
	EventMemberAdded         = "member_added"
	EventMemberRemoved       = "member_removed"
	EventMessageDeleted      = "message_deleted"
	EventError               = "error"
)

// BaseMessage is the base structure for all WebSocket messages
type BaseMessage struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
	ID      string      `json:"id,omitempty"` // client-generated UUID for ack
}

// ========== Client -> Server Payloads ==========

// PingPayload - empty payload for ping
type PingPayload struct{}

// JoinConversationPayload - join a conversation room
type JoinConversationPayload struct {
	ConversationID string `json:"conversation_id"`
}

// LeaveConversationPayload - leave a conversation room
type LeaveConversationPayload struct {
	ConversationID string `json:"conversation_id"`
}

// SendMessagePayload - send a message
type SendMessagePayload struct {
	ConversationID string `json:"conversation_id"`
	Body           string `json:"body"`
	Type           string `json:"type"`           // "TEXT" or "ATTACHMENT"
	ClientTempID   string `json:"client_temp_id"` // for optimistic UI
}

// TypingPayload - typing indicator
type TypingPayload struct {
	ConversationID string `json:"conversation_id"`
	IsTyping       bool   `json:"is_typing"`
}

// MarkReadPayload - mark messages as read
type MarkReadPayload struct {
	ConversationID string `json:"conversation_id"`
	LastMessageID  string `json:"last_message_id"`
}

// ========== Server -> Client Payloads ==========

// ConnectedPayload - welcome message on connection
type ConnectedPayload struct {
	UserID     string `json:"user_id"`
	ServerTime string `json:"server_time"`
}

// PongPayload - response to ping
type PongPayload struct {
	ServerTime string `json:"server_time"`
}

// SenderInfo - minimal sender info for message payloads
type SenderInfo struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// AttachmentInfo - attachment metadata
type AttachmentInfo struct {
	ID           string `json:"id"`
	FileType     string `json:"file_type"` // "IMAGE", "FILE", "VIDEO"
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Width        *int   `json:"width,omitempty"`
	Height       *int   `json:"height,omitempty"`
}

// MessageInfo - full message info
type MessageInfo struct {
	ID             string          `json:"id"`
	ConversationID string          `json:"conversation_id"`
	Sender         SenderInfo      `json:"sender"`
	Body           *string         `json:"body,omitempty"`
	Type           string          `json:"type"`   // "TEXT", "ATTACHMENT"
	Status         string          `json:"status"` // "SENT", "DELIVERED", "READ"
	Attachment     *AttachmentInfo `json:"attachment,omitempty"`
	CreatedAt      string          `json:"created_at"`
	DeletedAt      *string         `json:"deleted_at,omitempty"`
}

// NewMessagePayload - broadcast to all participants
type NewMessagePayload struct {
	Message      MessageInfo `json:"message"`
	ClientTempID *string     `json:"client_temp_id,omitempty"` // only for sender
}

// MessageSentPayload - ack to sender
type MessageSentPayload struct {
	ClientTempID string `json:"client_temp_id"`
	MessageID    string `json:"message_id"`
	CreatedAt    string `json:"created_at"`
	Status       string `json:"status"` // "SENT"
}

// UserTypingPayload - broadcast typing indicator
type UserTypingPayload struct {
	ConversationID string     `json:"conversation_id"`
	User           SenderInfo `json:"user"`
	IsTyping       bool       `json:"is_typing"`
}

// ReadReceiptPayload - read receipt for DM
type ReadReceiptPayload struct {
	ConversationID    string `json:"conversation_id"`
	ReaderID          string `json:"reader_id"`
	LastReadMessageID string `json:"last_read_message_id"`
	ReadAt            string `json:"read_at"`
}

// PresenceUpdatePayload - online/offline status
type PresenceUpdatePayload struct {
	UserID     string `json:"user_id"`
	Status     string `json:"status"` // "ONLINE", "OFFLINE"
	LastSeenAt string `json:"last_seen_at,omitempty"`
}

// ActorInfo - actor for notifications
type ActorInfo struct {
	ID          string  `json:"id"`
	Username    string  `json:"username,omitempty"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// NotificationInfo - notification data
type NotificationInfo struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"` // "FRIEND_REQUEST", "FRIEND_ACCEPTED", "GROUP_ADDED"
	Actor         ActorInfo `json:"actor"`
	ReferenceID   string    `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	PreviewText   string    `json:"preview_text"`
	CreatedAt     string    `json:"created_at"`
}

// NotificationNewPayload - new notification push
type NotificationNewPayload struct {
	Notification NotificationInfo `json:"notification"`
	UnreadCount  int              `json:"unread_count"`
}

// ConversationChanges - changes in conversation
type ConversationChanges struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdatedByInfo - who made the change
type UpdatedByInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ConversationUpdatedPayload - conversation metadata changed
type ConversationUpdatedPayload struct {
	ConversationID string              `json:"conversation_id"`
	Changes        ConversationChanges `json:"changes"`
	UpdatedBy      UpdatedByInfo       `json:"updated_by"`
}

// NewMemberInfo - new member data
type NewMemberInfo struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Role        string  `json:"role"` // "MEMBER", "ADMIN"
}

// MemberAddedPayload - member added to group
type MemberAddedPayload struct {
	ConversationID string        `json:"conversation_id"`
	NewMember      NewMemberInfo `json:"new_member"`
	AddedBy        UpdatedByInfo `json:"added_by"`
}

// MemberRemovedPayload - member removed from group
type MemberRemovedPayload struct {
	ConversationID string        `json:"conversation_id"`
	RemovedUserID  string        `json:"removed_user_id"`
	RemovedBy      UpdatedByInfo `json:"removed_by"`
	Reason         string        `json:"reason"` // "KICKED", "LEFT"
}

// MessageDeletedPayload - message soft deleted
type MessageDeletedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	DeletedBy      string `json:"deleted_by"`
	DeletedAt      string `json:"deleted_at"`
}

// ErrorPayload - error response
type ErrorPayload struct {
	Code     string `json:"code"` // "FORBIDDEN", "NOT_FOUND", etc.
	Message  string `json:"message"`
	RefEvent string `json:"ref_event,omitempty"` // which event caused the error
}

// ========== PostgreSQL NOTIFY Payloads ==========

// ChatEvent is the payload structure for pg_notify
type ChatEvent struct {
	Type           string      `json:"type"` // matches WS event name
	ConversationID string      `json:"conversation_id"`
	RecipientIDs   []string    `json:"recipient_ids"`        // nil = broadcast all participants
	Data           interface{} `json:"data"`                 // full event payload or message_id only
	MessageID      *string     `json:"message_id,omitempty"` // for large payload case
}
