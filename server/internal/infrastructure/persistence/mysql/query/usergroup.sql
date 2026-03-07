-- name: SelectUserGroupByUserId :many
SELECT *
FROM user_groups
WHERE user_id = ?;

-- name: SelectUserGroupByGroupId :many
SELECT *
FROM user_groups
WHERE group_id = ?;

-- name: CreateUserGroup :exec
INSERT INTO user_groups (user_id,
                         group_id,
                         created_at,
                         updated_at)
VALUES (?,
        ?,
        now(),
        now());

-- name: DeleteUserGroup :exec
DELETE
FROM user_groups
WHERE user_id = ?
  AND group_id = ?;

-- name: DeleteUserAllGroup :exec
DELETE
FROM user_groups
WHERE user_id = ?;

-- name: RemoveAllUsersFromGroup :exec
DELETE FROM user_groups WHERE group_id = ?;

-- name: IsUserInGroup :one
SELECT EXISTS(SELECT 1 FROM user_groups WHERE user_id = ? AND group_id = ?);
