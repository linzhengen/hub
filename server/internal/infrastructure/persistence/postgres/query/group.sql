-- name: SelectGroupById :one
SELECT *
FROM "groups"
WHERE id = $1 LIMIT 1;

-- name: SelectGroupForUpdate :one
SELECT *
FROM "groups"
WHERE id = $1 LIMIT 1 FOR
UPDATE;

-- name: UpdateGroup :exec
-- org_id is not updatable. Moving a group to another organization would carry
-- every member's access across a tenant boundary in one statement, which is a
-- migration rather than an edit.
UPDATE "groups"
SET name        = $1,
    description = $2,
    status      = $3,
    updated_at  = now()
WHERE id = $4;

-- name: CreateGroup :exec
INSERT INTO "groups" (id,
                      name,
                      status,
                      description,
                      org_id,
                      created_at,
                      updated_at)
VALUES ($1,
        $2,
        $3,
        $4,
        $5,
        now(),
        now());

-- name: DeleteGroup :exec
DELETE
FROM "groups"
WHERE id = $1;
