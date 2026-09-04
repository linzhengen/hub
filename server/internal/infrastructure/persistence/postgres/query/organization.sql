-- name: CreateOrganization :exec
INSERT INTO organizations (id,
                           name,
                           slug,
                           kind,
                           description,
                           status,
                           created_at,
                           updated_at)
VALUES (sqlc.arg(id),
        sqlc.arg(name),
        sqlc.arg(slug),
        sqlc.arg(kind),
        sqlc.arg(description),
        sqlc.arg(status),
        now(),
        now());

-- name: SelectOrganization :one
SELECT *
FROM organizations
WHERE id = $1
LIMIT 1;

-- The slug is how an operator and a URL name a tenant, so it is looked up as
-- often as the id is.
-- name: SelectOrganizationBySlug :one
SELECT *
FROM organizations
WHERE slug = $1
LIMIT 1;

-- name: UpdateOrganization :exec
-- The kind is not updatable. Changing an organization from BUSINESS to PLATFORM
-- would hand it every tenant's data in one statement, and changing it the other
-- way would revoke every administrator; neither is an edit, both are migrations.
UPDATE organizations
SET name        = sqlc.arg(name),
    slug        = sqlc.arg(slug),
    description = sqlc.arg(description),
    status      = sqlc.arg(status),
    updated_at  = now()
WHERE id = sqlc.arg(id);

-- name: DeleteOrganization :exec
DELETE
FROM organizations
WHERE id = $1;

-- The organizations a user can reach, by way of the groups they are in. There
-- is no membership table of its own: belonging to an organization is having a
-- group in it, which is the same edge an authorization decision reads.
-- name: SelectUserOrganizations :many
SELECT DISTINCT o.*
FROM organizations AS o
         INNER JOIN "groups" AS g ON g.org_id = o.id AND g.status = 'Active'
         INNER JOIN user_groups AS ug ON ug.group_id = g.id
WHERE ug.user_id = $1
  AND o.status = 'Active'
ORDER BY o.name;
