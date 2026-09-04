-- name: CreateAgent :exec
INSERT INTO agents (id,
                    org_id,
                    user_id,
                    name,
                    description,
                    client_id,
                    keycloak_id,
                    auth_method,
                    parent_agent_id,
                    created_by_user_id,
                    secret_rotated_at,
                    created_at,
                    updated_at)
VALUES (sqlc.arg(id),
        sqlc.arg(org_id),
        sqlc.arg(user_id),
        sqlc.arg(name),
        sqlc.arg(description),
        sqlc.arg(client_id),
        sqlc.arg(keycloak_id),
        sqlc.arg(auth_method),
        sqlc.narg(parent_agent_id),
        sqlc.arg(created_by_user_id),
        now(),
        now(),
        now());

-- name: SelectAgent :one
SELECT *
FROM agents
WHERE id = $1
LIMIT 1;

-- Every filter is optional and written as "unset or matching". org_ids is the
-- caller's reach rather than a filter they asked for: NULL is a platform grant,
-- which answers about every organization, and any other value is the set they
-- hold a live grant in. Listing every agent would tell one customer which
-- agents another runs.
-- name: ListAgents :many
SELECT *
FROM agents
WHERE (sqlc.narg(org_id)::uuid IS NULL OR org_id = sqlc.narg(org_id)::uuid)
  AND (sqlc.narg(parent_agent_id)::uuid IS NULL OR parent_agent_id = sqlc.narg(parent_agent_id)::uuid)
  AND (sqlc.narg(org_ids)::uuid[] IS NULL OR org_id = ANY (sqlc.narg(org_ids)::uuid[]))
ORDER BY name
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountAgents :one
SELECT count(*)
FROM agents
WHERE (sqlc.narg(org_id)::uuid IS NULL OR org_id = sqlc.narg(org_id)::uuid)
  AND (sqlc.narg(parent_agent_id)::uuid IS NULL OR parent_agent_id = sqlc.narg(parent_agent_id)::uuid)
  AND (sqlc.narg(org_ids)::uuid[] IS NULL OR org_id = ANY (sqlc.narg(org_ids)::uuid[]));

-- An agent with children cannot be removed: deleting it would leave their
-- credentials working with nothing in hub recording them. The count is asked
-- before anything is touched rather than left to the foreign key to refuse
-- half-way through the delete.
-- name: CountAgentChildren :one
SELECT count(*)
FROM agents
WHERE parent_agent_id = sqlc.arg(parent_agent_id)::uuid;

-- name: RecordAgentSecretRotation :exec
UPDATE agents
SET secret_rotated_at = now()
WHERE id = $1;

-- name: DeleteAgent :exec
DELETE
FROM agents
WHERE id = $1;
