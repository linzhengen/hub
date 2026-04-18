-- name: SelectUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: SelectUserForUpdate :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: UpdateUser :exec
UPDATE users
SET username   = $1,
    email      = $2,
    status     = $3,
    updated_at = now()
WHERE id = $4;

-- name: CreateUser :exec
INSERT INTO users (id,
                   username,
                   email,
                   status,
                   created_at,
                   updated_at)
VALUES ($1,
        $2,
        $3,
        $4,
        now(),
        now());

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;
