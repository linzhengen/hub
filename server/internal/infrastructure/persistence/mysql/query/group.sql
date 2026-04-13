-- name: SelectGroupById :one
SELECT *
FROM `groups`
WHERE id = ? LIMIT 1;

-- name: SelectGroupForUpdate :one
SELECT *
FROM `groups`
WHERE id = ? LIMIT 1 FOR
UPDATE;

-- name: UpdateGroup :exec
UPDATE `groups`
SET name        = ?,
    description = ?,
    status      = ?,
    updated_at  = now()
WHERE id = ?;

-- name: CreateGroup :exec
INSERT INTO `groups` (id,
                      name,
                      status,
                      description,
                      created_at,
                      updated_at)
VALUES (?,
        ?,
        ?,
        ?,
        now(),
        now());

-- name: DeleteGroup :exec
DELETE
FROM `groups`
WHERE id = ?;
