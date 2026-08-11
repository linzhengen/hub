-- name: CreateChatSession :one
INSERT INTO chat_sessions (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: SelectChatSessionById :one
SELECT *
FROM chat_sessions
WHERE id = $1
LIMIT 1;

-- name: SelectChatSessionsByUserId :many
SELECT *
FROM chat_sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteChatSession :exec
DELETE
FROM chat_sessions
WHERE id = $1;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (session_id, role, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: SelectChatMessagesBySessionId :many
SELECT *
FROM chat_messages
WHERE session_id = $1
ORDER BY created_at ASC;
