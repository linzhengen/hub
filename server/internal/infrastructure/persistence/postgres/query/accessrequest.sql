-- name: CreateAccessRequest :exec
INSERT INTO access_requests (id,
                             requester_user_id,
                             subject_user_id,
                             group_id,
                             reason,
                             requested_until,
                             status,
                             origin,
                             session_id,
                             created_at,
                             updated_at)
VALUES (sqlc.arg(id),
        sqlc.arg(requester_user_id),
        sqlc.arg(subject_user_id),
        sqlc.arg(group_id),
        sqlc.arg(reason),
        sqlc.narg(requested_until),
        sqlc.arg(status),
        sqlc.arg(origin),
        sqlc.narg(session_id),
        now(),
        now());

-- name: SelectAccessRequest :one
SELECT *
FROM access_requests
WHERE id = $1
LIMIT 1;

-- Deciding is conditional on the request still being pending, so two decisions
-- racing cannot both grant the access: the second updates no row.
-- name: DecideAccessRequest :one
UPDATE access_requests
SET status             = sqlc.arg(status),
    decided_by_user_id = sqlc.arg(decided_by_user_id),
    decided_at         = sqlc.arg(decided_at),
    decision_comment   = sqlc.arg(decision_comment)
WHERE id = sqlc.arg(id)
  AND status = 'Pending'
RETURNING *;

-- Every filter is optional and written as "unset or matching", so one query
-- serves the approver's queue ("pending"), the requester's own list ("mine")
-- and a group's history without three near-identical statements.
-- name: ListAccessRequests :many
SELECT *
FROM access_requests
WHERE (sqlc.narg(requester_user_id)::uuid IS NULL OR requester_user_id = sqlc.narg(requester_user_id)::uuid)
  AND (sqlc.narg(subject_user_id)::uuid IS NULL OR subject_user_id = sqlc.narg(subject_user_id)::uuid)
  AND (sqlc.narg(group_id)::uuid IS NULL OR group_id = sqlc.narg(group_id)::uuid)
  AND (sqlc.narg(status)::varchar IS NULL OR status = sqlc.narg(status)::varchar)
-- Newest first: a queue is read from the top.
ORDER BY created_at DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountAccessRequests :one
SELECT count(*)
FROM access_requests
WHERE (sqlc.narg(requester_user_id)::uuid IS NULL OR requester_user_id = sqlc.narg(requester_user_id)::uuid)
  AND (sqlc.narg(subject_user_id)::uuid IS NULL OR subject_user_id = sqlc.narg(subject_user_id)::uuid)
  AND (sqlc.narg(group_id)::uuid IS NULL OR group_id = sqlc.narg(group_id)::uuid)
  AND (sqlc.narg(status)::varchar IS NULL OR status = sqlc.narg(status)::varchar);
