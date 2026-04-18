-- name: SelectPermissionById :one
SELECT *
FROM permissions
WHERE id = $1 LIMIT 1;

-- name: SelectPermissionForUpdate :one
SELECT *
FROM permissions
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: SelectPermissionByResourceId :many
SELECT *
FROM permissions
WHERE resource_id = $1;

-- name: UpdatePermission :exec
UPDATE permissions
SET verb        = $1,
    resource_id = $2,
    description = $3,
    updated_at  = now()
WHERE id = $4;

-- name: CreatePermission :exec
INSERT INTO permissions (id,
                         verb,
                         resource_id,
                         description,
                         created_at,
                         updated_at)
VALUES ($1,
        $2,
        $3,
        $4,
        now(),
        now());

-- name: DeletePermissions :exec
DELETE
FROM "permissions"
WHERE id = $1;
