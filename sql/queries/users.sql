-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password, nickname, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING *;

-- name: NukeUsers :exec
DELETE FROM users WHERE true;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: FindUserById :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_password = $2, nickname = $3, updated_at = NOW()
WHERE id = $3
RETURNING *;

