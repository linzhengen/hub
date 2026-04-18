-- name: SelectRolePermissionByRoleId :many
SELECT *
FROM role_permissions
WHERE role_id = $1;

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id,
                              permission_id,
                              created_at,
                              updated_at)
VALUES ($1,
        $2,
        now(),
        now());

-- name: RemovePermissionFromRole :exec
DELETE
FROM role_permissions
WHERE role_id = $1
  AND permission_id = $2;

-- name: DeleteRoleAllPermission :exec
DELETE
FROM role_permissions
WHERE role_id = $1;

-- name: IsPermissionInRole :one
SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = $1 AND permission_id = $2);
