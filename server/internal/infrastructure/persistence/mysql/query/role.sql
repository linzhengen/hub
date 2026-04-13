-- name: SelectRoleById :one
SELECT *
FROM roles
WHERE id = ? LIMIT 1;

-- name: SelectRoleForUpdate :one
SELECT *
FROM roles
WHERE id = ? LIMIT 1 FOR UPDATE;

-- name: UpdateRole :exec
UPDATE roles
SET name        = ?,
    description = ?,
    updated_at  = now()
WHERE id = ?;

-- name: CreateRole :exec
INSERT INTO roles (id,
                    name,
                    description,
                    created_at,
                    updated_at)
VALUES (?,
        ?,
        ?,
        now(),
        now());

-- name: DeleteRole :exec
DELETE
FROM `roles`
WHERE id = ?;
