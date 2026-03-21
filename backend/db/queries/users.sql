-- name: CreateUser :one
INSERT INTO users (id, email, username, display_name, password_hash, status)
VALUES ($1, $2, $3, $4, $5, 'ONLINE')
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: FindUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $1, last_seen_at = NOW() WHERE id = $2;
