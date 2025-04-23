-- name: CreateChatroom :one
INSERT INTO chatrooms (id, type, name, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: DeleteChatroom :exec
DELETE FROM chatrooms WHERE id = $1;

-- name: FindUsersChatrooms :many
SELECT cr.* FROM chatrooms AS cr 
LEFT JOIN chatrooms_participants AS cp ON cr.id = cp.chatroom_id
WHERE cp.participant_id = $1;

-- name: FindChatRoomById :one
SELECT * FROM chatrooms WHERE id = $1;

-- name: UpdateChatroom :one
UPDATE chatrooms
SET type = $1, name = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: ChatroomAddParticipant :exec
INSERT INTO chatrooms_participants(chatroom_id, participant_id)
VALUES ($1, $2);

-- name: ChatroomRemoveParticipant :exec
DELETE FROM chatrooms_participants
WHERE chatroom_id = $1 AND participant_id = $2;
