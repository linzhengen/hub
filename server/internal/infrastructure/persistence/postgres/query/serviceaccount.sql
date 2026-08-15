-- name: CreateServiceAccount :exec
INSERT INTO service_accounts (id,
                              user_id,
                              name,
                              description,
                              client_id,
                              keycloak_id,
                              created_by_user_id,
                              created_at,
                              updated_at)
VALUES (sqlc.arg(id),
        sqlc.arg(user_id),
        sqlc.arg(name),
        sqlc.arg(description),
        sqlc.arg(client_id),
        sqlc.arg(keycloak_id),
        sqlc.arg(created_by_user_id),
        now(),
        now());

-- name: SelectServiceAccount :one
SELECT *
FROM service_accounts
WHERE id = $1
LIMIT 1;

-- The listing is not filtered. A deployment has a handful of machines, not a
-- directory of them, and "who created it" is a column to read rather than a
-- question to ask the database.
-- name: ListServiceAccounts :many
SELECT *
FROM service_accounts
ORDER BY name
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountServiceAccounts :one
SELECT count(*)
FROM service_accounts;

-- name: DeleteServiceAccount :exec
DELETE
FROM service_accounts
WHERE id = $1;
