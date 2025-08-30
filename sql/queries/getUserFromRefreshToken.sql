-- name: GetUserFromRefreshToken :one
SELECT
    users.id,
    users.email,
    users.created_at,
    users.updated_at,
    users.hashed_password,
    refresh_tokens.token
FROM refresh_tokens
INNER JOIN users ON refresh_tokens.user_id = users.id
WHERE token = $1 AND revoked_at IS NULL AND expires_at > Now();

