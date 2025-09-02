-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = Now(), updated_at = Now()
WHERE token = $1;


