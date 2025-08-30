-- name: RefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
	$1,
	Now(),
	Now(),
	$2,
	Now() + INTERVAL '60 days',
	NULL
)
RETURNING *;




