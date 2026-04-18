-- name: SelectResourceById :one
SELECT *
FROM resources
WHERE id = $1 LIMIT 1;

-- name: SelectResourceByIdentifier :one
SELECT *
FROM resources
WHERE identifier = $1 LIMIT 1;

-- name: SelectResourceForUpdate :one
SELECT *
FROM resources
WHERE id = $1 LIMIT 1 FOR
UPDATE;

-- name: UpdateResource :exec
UPDATE resources
SET parent_id     = $1,
    name          = $2,
    identifier    = $3,
    type          = $4,
    path          = $5,
    component     = $6,
    display_order = $7,
    description   = $8,
    metadata      = $9,
    status        = $10,
    updated_at    = now()
WHERE id = $11;

-- name: CreateResource :exec
INSERT INTO resources (id,
                       parent_id,
                       name,
                       identifier,
                       type,
                       path,
                       component,
                       display_order,
                       description,
                       metadata,
                       status,
                       created_at,
                       updated_at)
VALUES ($1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        now(),
        now());

-- name: DeleteResource :exec
DELETE
FROM "resources"
WHERE id = $1;
