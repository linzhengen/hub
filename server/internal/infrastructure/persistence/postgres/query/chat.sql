-- name: CreateChatSession :one
INSERT INTO chat_sessions (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- Scoping every lookup by user_id keeps a session private to its owner even if
-- a caller reaches the repository without going through the use case, which is
-- where the check used to live alone. A session belonging to someone else
-- returns no row, so it is indistinguishable from one that does not exist.
-- name: SelectChatSessionById :one
SELECT *
FROM chat_sessions
WHERE id = $1
  AND user_id = $2
LIMIT 1;

-- name: SelectChatSessionsByUserId :many
SELECT *
FROM chat_sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteChatSession :exec
DELETE
FROM chat_sessions
WHERE id = $1
  AND user_id = $2;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (session_id, role, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: SelectChatMessagesBySessionId :many
SELECT m.*
FROM chat_messages m
         JOIN chat_sessions s ON s.id = m.session_id
WHERE m.session_id = $1
  AND s.user_id = $2
ORDER BY m.created_at ASC;

-- name: CreateChatToolProposal :one
INSERT INTO chat_tool_proposals (session_id, tool_name, arguments, continuation)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- Scoped by user_id through the session, for the same reason every other chat
-- query is: a proposal belonging to someone else is absent, not refused.
-- name: SelectChatToolProposalById :one
SELECT p.*
FROM chat_tool_proposals p
         JOIN chat_sessions s ON s.id = p.session_id
WHERE p.id = $1
  AND s.user_id = $2
LIMIT 1;

-- Deciding is conditional on the proposal still being pending, so two decisions
-- racing cannot both run the change: the second updates no row.
-- name: DecideChatToolProposal :one
UPDATE chat_tool_proposals
SET status     = $2,
    decided_at = now()
WHERE id = $1
  AND status = 'pending'
RETURNING *;

-- Adding rather than setting, so two streams on one session both count.
-- name: AddChatSessionTokens :exec
UPDATE chat_sessions
SET tokens_used = tokens_used + $2
WHERE id = $1;
