-- name: SelectPermissionById :one
SELECT *
FROM permissions
WHERE id = ? LIMIT 1;

-- name: SelectPermissionForUpdate :one
SELECT *
FROM permissions
WHERE id = ? LIMIT 1 FOR UPDATE;

-- name: SelectPermissionByResourceId :many
SELECT *
FROM permissions
WHERE resource_id = ?;

-- name: UpdatePermission :exec
UPDATE permissions
SET verb        = ?,
    resource_id = ?,
    description = ?,
    updated_at  = now()
WHERE id = ?;

-- name: CreatePermission :exec
INSERT INTO permissions (id,
                         verb,
                         resource_id,
                         description,
                         created_at,
                         updated_at)
VALUES (?,
        ?,
        ?,
        ?,
        now(),
        now());

-- name: DeletePermissions :exec
DELETE
FROM `permissions`
WHERE id = ?;
