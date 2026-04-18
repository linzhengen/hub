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
                      created_at,
                      updated_at)
VALUES ($1,
        $2,
        $3,
        $4,
        now(),
        now());

-- name: DeleteGroup :exec
DELETE
FROM "groups"
WHERE id = $1;
