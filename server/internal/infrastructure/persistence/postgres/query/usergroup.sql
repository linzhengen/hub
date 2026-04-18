-- name: SelectUserGroupByUserId :many
SELECT *
FROM user_groups
WHERE user_id = $1;

-- name: SelectUserGroupByGroupId :many
SELECT *
FROM user_groups
WHERE group_id = $1;

-- name: CreateUserGroup :exec
INSERT INTO user_groups (user_id,
                         group_id,
                         created_at,
                         updated_at)
VALUES ($1,
        $2,
        now(),
        now());

-- name: DeleteUserGroup :exec
DELETE
FROM user_groups
WHERE user_id = $1
  AND group_id = $2;

-- name: DeleteUserAllGroup :exec
DELETE
FROM user_groups
WHERE user_id = $1;

-- name: RemoveAllUsersFromGroup :exec
DELETE FROM user_groups WHERE group_id = $1;

-- name: IsUserInGroup :one
SELECT EXISTS(SELECT 1 FROM user_groups WHERE user_id = $1 AND group_id = $2);
