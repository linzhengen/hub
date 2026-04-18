-- name: SelectGroupRoleByGroupId :many
SELECT *
FROM group_roles
WHERE group_id = $1;

-- name: CreateGroupRole :exec
INSERT INTO group_roles (group_id,
                         role_id,
                         created_at,
                         updated_at)
VALUES ($1,
        $2,
        now(),
        now());

-- name: DeleteGroupRole :exec
DELETE
FROM group_roles
WHERE group_id = $1
  AND role_id = $2;

-- name: DeleteGroupAllRole :exec
DELETE
FROM group_roles
WHERE group_id = $1;
