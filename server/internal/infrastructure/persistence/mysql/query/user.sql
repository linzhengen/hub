-- name: SelectUserById :one
SELECT *
FROM users
WHERE id = ? LIMIT 1;

-- name: SelectUserForUpdate :one
SELECT *
FROM users
WHERE id = ? LIMIT 1 FOR UPDATE;

-- name: UpdateUser :exec
UPDATE users
SET username   = ?,
    email      = ?,
    status     = ?,
    updated_at = now()
WHERE id = ?;

-- name: CreateUser :exec
INSERT INTO users (id,
                   username,
                   email,
                   status,
                   created_at,
                   updated_at)
VALUES (?,
        ?,
        ?,
        ?,
        now(),
        now());

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = ?;
