-- name: SelectRolePermissionByRoleId :many
SELECT *
FROM role_permissions
WHERE role_id = ?;

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id,
                              permission_id,
                              created_at,
                              updated_at)
VALUES (?,
        ?,
        now(),
        now());

-- name: RemovePermissionFromRole :exec
DELETE
FROM role_permissions
WHERE role_id = ?
  AND permission_id = ?;

-- name: DeleteRoleAllPermission :exec
DELETE
FROM role_permissions
WHERE role_id = ?;

-- name: IsPermissionInRole :one
SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = ? AND permission_id = ?);
