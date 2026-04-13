-- name: SelectGroupRoleByGroupId :many
SELECT *
FROM group_roles
WHERE group_id = ?;

-- name: CreateGroupRole :exec
INSERT INTO group_roles (group_id,
                         role_id,
                         created_at,
                         updated_at)
VALUES (?,
        ?,
        now(),
        now());

-- name: DeleteGroupRole :exec
DELETE
FROM group_roles
WHERE group_id = ?
  AND role_id = ?;

-- name: DeleteGroupAllRole :exec
DELETE
FROM group_roles
WHERE group_id = ?;
