-- name: CreateNotification :one
INSERT INTO notifications (id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, FALSE, NOW())
RETURNING id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at, read_at;

-- name: GetNotificationByID :one
SELECT id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at, read_at
FROM notifications
WHERE id = $1;

-- name: GetNotificationByReference :one
SELECT id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at, read_at
FROM notifications
WHERE recipient_id = $1 AND type = $2 AND reference_id = $3
LIMIT 1;

-- name: ListNotificationsByRecipient :many
SELECT
    n.id,
    n.recipient_id,
    n.actor_id,
    u.username AS actor_username,
    u.display_name AS actor_display_name,
    u.avatar_url AS actor_avatar_url,
    n.type,
    n.reference_id,
    n.reference_type,
    n.is_read,
    n.created_at,
    n.read_at
FROM notifications n
JOIN users u ON u.id = n.actor_id
WHERE n.recipient_id = $1
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :one
UPDATE notifications
SET is_read = TRUE, read_at = NOW()
WHERE id = $1 AND recipient_id = $2
RETURNING id, recipient_id, actor_id, type, reference_id, reference_type, is_read, created_at, read_at;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET is_read = TRUE, read_at = NOW()
WHERE recipient_id = $1 AND is_read = FALSE;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) as count
FROM notifications
WHERE recipient_id = $1 AND is_read = FALSE;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: DeleteNotificationByReference :exec
DELETE FROM notifications
WHERE recipient_id = $1 AND type = $2 AND reference_id = $3;
