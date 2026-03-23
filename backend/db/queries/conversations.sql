-- name: ListConversationsByUser :many
-- Lấy danh sách conversations của user với unread count và last message
-- Sử dụng cursor pagination dựa trên last_activity_at
SELECT
    c.id,
    c.type,
    c.name,
    c.avatar_url,
    c.last_activity_at,
    m_last.id AS last_message_id,
    m_last.body AS last_message_body,
    m_last.type AS last_message_type,
    m_last.created_at AS last_message_created_at,
    m_last.sender_id AS last_message_sender_id,
    u_last.display_name AS last_message_sender_name,
    u_last.avatar_url AS last_message_sender_avatar,
    COUNT(m_unread.id) AS unread_count,
    -- Thông tin other_user cho DM (sẽ được xử lý ở application layer)
    CASE 
        WHEN c.type = 'DM' THEN (
            SELECT cp2.user_id 
            FROM conversation_participants cp2 
            WHERE cp2.conversation_id = c.id 
            AND cp2.user_id != $1 
            AND cp2.left_at IS NULL
            LIMIT 1
        )
    END AS other_user_id
FROM conversation_participants cp
JOIN conversations c ON c.id = cp.conversation_id
LEFT JOIN messages m_last ON m_last.id = c.last_message_id
LEFT JOIN users u_last ON u_last.id = m_last.sender_id
LEFT JOIN messages m_unread ON 
    m_unread.conversation_id = c.id
    AND m_unread.deleted_at IS NULL
    AND m_unread.sender_id != $1
    AND m_unread.created_at > COALESCE(cp.last_read_at, '1970-01-01')
WHERE cp.user_id = $1
    AND cp.left_at IS NULL
    AND ($2::timestamptz IS NULL OR c.last_activity_at < $2)
    AND ($3::uuid IS NULL OR c.id < $3)
GROUP BY c.id, m_last.id, m_last.body, m_last.type, m_last.created_at, 
         m_last.sender_id, u_last.display_name, u_last.avatar_url
ORDER BY c.last_activity_at DESC, c.id DESC
LIMIT $4;

-- name: FindDMConversation :one
-- Tìm DM conversation giữa 2 users (dùng cho reuse khi re-friend)
SELECT c.id
FROM conversations c
JOIN conversation_participants cp1 ON cp1.conversation_id = c.id
JOIN conversation_participants cp2 ON cp2.conversation_id = c.id
WHERE c.type = 'DM'
    AND cp1.user_id = $1
    AND cp2.user_id = $2
    AND cp1.left_at IS NULL
    AND cp2.left_at IS NULL
LIMIT 1;

-- name: CreateConversation :one
-- Tạo conversation mới (DM hoặc GROUP)
INSERT INTO conversations (id, type, name, last_activity_at)
VALUES ($1, $2, $3, NOW())
RETURNING id, type, name, avatar_url, last_activity_at, created_at;

-- name: CreateParticipant :one
-- Thêm participant vào conversation
INSERT INTO conversation_participants (id, conversation_id, user_id, role, joined_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, conversation_id, user_id, role, last_read_at, joined_at;

-- name: UpdateLastActivity :exec
-- Cập nhật last_activity_at khi có hoạt động mới
UPDATE conversations
SET last_activity_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateLastMessage :exec
-- Cập nhật last_message_id sau khi gửi tin nhắn
UPDATE conversations
SET last_message_id = $2,
    last_activity_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: GetConversationByID :one
-- Lấy chi tiết conversation theo ID
SELECT c.id, c.type, c.name, c.avatar_url, c.last_activity_at, c.created_at
FROM conversations c
WHERE c.id = $1;

-- name: GetConversationParticipants :many
-- Lấy danh sách participants của conversation
SELECT 
    cp.id,
    cp.user_id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.status,
    cp.role,
    cp.joined_at
FROM conversation_participants cp
JOIN users u ON u.id = cp.user_id
WHERE cp.conversation_id = $1
    AND cp.left_at IS NULL
ORDER BY cp.joined_at;

-- name: IsConversationMember :one
-- Kiểm tra user có phải member của conversation không
SELECT EXISTS(
    SELECT 1 FROM conversation_participants
    WHERE conversation_id = $1 
    AND user_id = $2 
    AND left_at IS NULL
) AS is_member;

-- name: IsGroupAdmin :one
-- Kiểm tra user có phải admin của group không
SELECT EXISTS(
    SELECT 1 FROM conversation_participants
    WHERE conversation_id = $1 
    AND user_id = $2 
    AND role = 'ADMIN'
    AND left_at IS NULL
) AS is_admin;

-- name: UpdateLastRead :one
-- Cập nhật last_read_message_id và last_read_at
UPDATE conversation_participants
SET last_read_message_id = $3,
    last_read_at = NOW()
WHERE conversation_id = $1 
    AND user_id = $2
RETURNING id, last_read_message_id, last_read_at;

-- name: GetFriendshipStatus :one
-- Kiểm tra status friendship giữa 2 users
SELECT status FROM friendships
WHERE ((requester_id = $1 AND addressee_id = $2)
    OR (requester_id = $2 AND addressee_id = $1))
    AND status = 'ACCEPTED'
LIMIT 1;

-- name: UpdateConversation :one
-- Cập nhật tên hoặc avatar của conversation
UPDATE conversations
SET name = COALESCE($2, name),
    avatar_url = COALESCE($3, avatar_url),
    avatar_cloudinary_id = COALESCE($4, avatar_cloudinary_id),
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, avatar_url, updated_at;
