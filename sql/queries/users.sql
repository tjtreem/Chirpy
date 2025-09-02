-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;


-- name: UpgradeUserToChirpyRed :execresult
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;


