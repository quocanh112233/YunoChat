-- name: CreateFriendship :one
INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, 'PENDING', NOW(), NOW())
RETURNING id, requester_id, addressee_id, status, created_at, updated_at;

-- name: GetFriendshipByID :one
SELECT id, requester_id, addressee_id, status, created_at, updated_at
FROM friendships
WHERE id = $1;

-- name: GetFriendshipBetweenUsers :one
SELECT id, requester_id, addressee_id, status, created_at, updated_at
FROM friendships
WHERE (requester_id = $1 AND addressee_id = $2)
   OR (requester_id = $2 AND addressee_id = $1)
LIMIT 1;

-- name: ListFriendsByUser :many
SELECT
    f.id AS friendship_id,
    CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END AS friend_id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.status,
    u.last_seen_at,
    f.created_at AS became_friends_at
FROM friendships f
JOIN users u ON u.id = CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END
WHERE (f.requester_id = $1 OR f.addressee_id = $1)
  AND f.status = 'ACCEPTED'
ORDER BY f.updated_at DESC;

-- name: ListPendingRequestsReceived :many
SELECT
    f.id AS request_id,
    f.requester_id AS from_user_id,
    u.username,
    u.display_name,
    u.avatar_url,
    f.created_at AS requested_at
FROM friendships f
JOIN users u ON u.id = f.requester_id
WHERE f.addressee_id = $1
  AND f.status = 'PENDING'
ORDER BY f.created_at DESC;

-- name: ListPendingRequestsSent :many
SELECT
    f.id AS request_id,
    f.addressee_id AS to_user_id,
    u.username,
    u.display_name,
    u.avatar_url,
    f.created_at AS requested_at
FROM friendships f
JOIN users u ON u.id = f.addressee_id
WHERE f.requester_id = $1
  AND f.status = 'PENDING'
ORDER BY f.created_at DESC;

-- name: UpdateFriendshipStatus :one
UPDATE friendships
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, requester_id, addressee_id, status, created_at, updated_at;

-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE id = $1;

-- name: FindDMConversationBetweenUsers :one
SELECT c.id
FROM conversations c
JOIN conversation_participants cp1 ON cp1.conversation_id = c.id AND cp1.user_id = $1 AND cp1.left_at IS NULL
JOIN conversation_participants cp2 ON cp2.conversation_id = c.id AND cp2.user_id = $2 AND cp2.left_at IS NULL
WHERE c.type = 'DM'
LIMIT 1;

-- name: SearchUsersWithRelationship :many
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.status,
    CASE
        WHEN f_accepted.id IS NOT NULL THEN 'ACCEPTED'
        WHEN f_pending_sent.id IS NOT NULL THEN 'PENDING_SENT'
        WHEN f_pending_received.id IS NOT NULL THEN 'PENDING_RECEIVED'
        ELSE 'NONE'
    END AS relationship
FROM users u
LEFT JOIN friendships f_accepted ON (
    (f_accepted.requester_id = $1 AND f_accepted.addressee_id = u.id) OR
    (f_accepted.addressee_id = $1 AND f_accepted.requester_id = u.id)
) AND f_accepted.status = 'ACCEPTED'
LEFT JOIN friendships f_pending_sent ON (
    f_pending_sent.requester_id = $1 AND f_pending_sent.addressee_id = u.id
) AND f_pending_sent.status = 'PENDING'
LEFT JOIN friendships f_pending_received ON (
    f_pending_received.addressee_id = $1 AND f_pending_received.requester_id = u.id
) AND f_pending_received.status = 'PENDING'
WHERE u.id <> $1
  AND (u.username ILIKE '%' || $2 || '%' OR u.display_name ILIKE '%' || $2 || '%')
ORDER BY u.display_name
LIMIT $3;
