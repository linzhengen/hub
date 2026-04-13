-- name: SelectResourceById :one
SELECT *
FROM resources
WHERE id = ? LIMIT 1;

-- name: SelectResourceByIdentifier :one
SELECT *
FROM resources
WHERE identifier = ? LIMIT 1;

-- name: SelectResourceForUpdate :one
SELECT *
FROM resources
WHERE id = ? LIMIT 1 FOR
UPDATE;

-- name: UpdateResource :exec
UPDATE resources
SET parent_id     = ?,
    name          = ?,
    identifier    = ?,
    type          = ?,
    path          = ?,
    component     = ?,
    display_order = ?,
    description   = ?,
    metadata      = ?,
    status        = ?,
    updated_at    = now()
WHERE id = ?;

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
VALUES (?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        now(),
        now());

-- name: DeleteResource :exec
DELETE
FROM `resources`
WHERE id = ?;
