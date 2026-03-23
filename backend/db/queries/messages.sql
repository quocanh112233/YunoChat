-- name: CreateMessage :one
-- Tạo message mới (TEXT hoặc ATTACHMENT)
INSERT INTO messages (id, conversation_id, sender_id, body, type, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id, conversation_id, sender_id, body, type, status, created_at, deleted_at;

-- name: ListMessages :many
-- Lấy danh sách messages với cursor pagination
-- Sử dụng composite cursor (created_at, id) để đảm bảo deterministic ordering
SELECT 
    m.id,
    m.conversation_id,
    m.sender_id,
    u.display_name AS sender_display_name,
    u.username AS sender_username,
    u.avatar_url AS sender_avatar_url,
    m.body,
    m.type,
    m.status,
    m.created_at,
    m.deleted_at,
    -- Attachment info (nếu có)
    a.id AS attachment_id,
    a.storage_type,
    a.file_type,
    a.url,
    a.thumbnail_url,
    a.original_name,
    a.mime_type,
    a.size_bytes,
    a.width,
    a.height,
    a.duration_secs
FROM messages m
JOIN users u ON u.id = m.sender_id
LEFT JOIN attachments a ON a.message_id = m.id
WHERE m.conversation_id = $1
    AND m.deleted_at IS NULL
    AND (
        $2::timestamptz IS NULL OR $3::uuid IS NULL
        OR (m.created_at, m.id) < ($2, $3)
    )
ORDER BY m.created_at DESC, m.id DESC
LIMIT $4;

-- name: SoftDeleteMessage :execrows
-- Soft delete message: chỉ sender mới được xóa tin nhắn của mình
-- Set deleted_at=NOW(), body=NULL (theo docs/7 phần 7)
UPDATE messages
SET deleted_at = NOW(),
    body = NULL,
    updated_at = NOW()
WHERE id = $1 
    AND sender_id = $2
    AND deleted_at IS NULL;

-- name: GetMessageByID :one
-- Lấy message theo ID (không bao gồm deleted)
SELECT 
    m.id,
    m.conversation_id,
    m.sender_id,
    u.display_name AS sender_display_name,
    u.username AS sender_username,
    u.avatar_url AS sender_avatar_url,
    m.body,
    m.type,
    m.status,
    m.created_at,
    m.deleted_at,
    a.id AS attachment_id,
    a.storage_type,
    a.file_type,
    a.url,
    a.thumbnail_url,
    a.original_name,
    a.mime_type,
    a.size_bytes,
    a.width,
    a.height,
    a.duration_secs
FROM messages m
JOIN users u ON u.id = m.sender_id
LEFT JOIN attachments a ON a.message_id = m.id
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: UpdateMessageStatus :exec
-- Cập nhật status của message (SENT -> DELIVERED -> READ)
-- Chỉ dùng cho DM
UPDATE messages
SET status = $2,
    updated_at = NOW()
WHERE conversation_id = $1
    AND sender_id != $3
    AND status < $2;  -- SENT < DELIVERED < READ

-- name: CreateAttachment :one
-- Tạo attachment record liên kết với message
INSERT INTO attachments (
    id, message_id, storage_type, file_type, url, thumbnail_url,
    original_name, mime_type, size_bytes, width, height, duration_secs
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, message_id, storage_type, file_type, url, thumbnail_url,
    original_name, mime_type, size_bytes, width, height, duration_secs, created_at;

-- name: GetUnreadCount :one
-- Đếm số tin nhắn chưa đọc trong conversation cho user cụ thể
SELECT COUNT(*)
FROM messages m
WHERE m.conversation_id = $1
    AND m.deleted_at IS NULL
    AND m.sender_id != $2
    AND m.created_at > COALESCE(
        (SELECT last_read_at FROM conversation_participants 
         WHERE conversation_id = $1 AND user_id = $2),
        '1970-01-01'
    );

-- name: GetLatestMessageID :one
-- Lấy ID của tin nhắn mới nhất trong conversation
SELECT id FROM messages
WHERE conversation_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1;
