-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: FindRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 AND is_revoked = false AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET is_revoked = true WHERE id = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1;
