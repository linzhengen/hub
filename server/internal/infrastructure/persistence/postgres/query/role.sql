-- name: SelectRoleById :one
SELECT *
FROM roles
WHERE id = $1 LIMIT 1;

-- name: SelectRoleForUpdate :one
SELECT *
FROM roles
WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: UpdateRole :exec
UPDATE roles
SET name        = $1,
    description = $2,
    updated_at  = now()
WHERE id = $3;

-- name: CreateRole :exec
INSERT INTO roles (id,
                    name,
                    description,
                    created_at,
                    updated_at)
VALUES ($1,
        $2,
        $3,
        now(),
        now());

-- name: DeleteRole :exec
DELETE
FROM "roles"
WHERE id = $1;
