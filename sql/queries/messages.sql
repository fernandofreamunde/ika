-- name: CreateMessage :one
INSERT INTO messages (id, type, content, author_id, chatroom_id, sent_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;

-- name: FindMessagesByRoomById :one
SELECT * 
FROM messages
WHERE chatroom_id = $1
ORDER BY sent_at DESC;

-- name: UpdateMessage :one
UPDATE messages
SET content = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

