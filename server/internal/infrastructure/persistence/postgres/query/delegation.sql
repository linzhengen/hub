-- name: CreateDelegation :exec
INSERT INTO delegations (id,
                         agent_id,
                         principal_user_id,
                         granted_by_user_id,
                         org_id,
                         reason,
                         max_depth,
                         expires_at,
                         created_at,
                         updated_at)
VALUES (sqlc.arg(id),
        sqlc.arg(agent_id),
        sqlc.arg(principal_user_id),
        sqlc.arg(granted_by_user_id),
        sqlc.arg(org_id),
        sqlc.arg(reason),
        sqlc.arg(max_depth),
        sqlc.narg(expires_at),
        now(),
        now());

-- name: AddPermissionToDelegation :exec
INSERT INTO delegation_permissions (delegation_id, permission_id, created_at, updated_at)
VALUES (sqlc.arg(delegation_id), sqlc.arg(permission_id), now(), now())
ON CONFLICT (delegation_id, permission_id) DO NOTHING;

-- name: SelectDelegation :one
SELECT *
FROM delegations
WHERE id = $1
LIMIT 1;

-- name: SelectDelegationPermissions :many
SELECT *
FROM delegation_permissions
WHERE delegation_id = $1
ORDER BY permission_id;

-- Every filter is optional and written as "unset or matching". Two of them are
-- not filters the caller asked for:
--
--   org_ids     the organizations the caller holds a live grant in. NULL is a
--               platform grant, which answers about every organization.
--   self_user_id the caller themselves, always visible whatever org_ids says -
--               a person must be able to see, and therefore revoke, what they
--               granted, even if the agent lives somewhere they cannot
--               otherwise read.
--
-- `revoked_at` is filtered here because revoking is a write: the revision
-- counter moves and caches drop. `expires_at` is deliberately *not* filtered -
-- an expiry is the one change nobody writes, so a query that hid lapsed rows
-- would let a warm cache keep serving one.
-- name: ListDelegations :many
SELECT *
FROM delegations
WHERE (sqlc.narg(agent_id)::uuid IS NULL OR agent_id = sqlc.narg(agent_id)::uuid)
  AND (sqlc.narg(principal_user_id)::uuid IS NULL OR principal_user_id = sqlc.narg(principal_user_id)::uuid)
  AND (sqlc.arg(include_revoked)::bool OR revoked_at IS NULL)
  AND (sqlc.narg(org_ids)::uuid[] IS NULL
       OR org_id = ANY (sqlc.narg(org_ids)::uuid[])
       OR principal_user_id = sqlc.narg(self_user_id)::uuid)
ORDER BY created_at DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountDelegations :one
SELECT count(*)
FROM delegations
WHERE (sqlc.narg(agent_id)::uuid IS NULL OR agent_id = sqlc.narg(agent_id)::uuid)
  AND (sqlc.narg(principal_user_id)::uuid IS NULL OR principal_user_id = sqlc.narg(principal_user_id)::uuid)
  AND (sqlc.arg(include_revoked)::bool OR revoked_at IS NULL)
  AND (sqlc.narg(org_ids)::uuid[] IS NULL
       OR org_id = ANY (sqlc.narg(org_ids)::uuid[])
       OR principal_user_id = sqlc.narg(self_user_id)::uuid);

-- Matching only a live row is what makes a repeated revocation not a second
-- event: the second statement changes nothing and the caller is told so.
-- name: RevokeDelegation :execrows
UPDATE delegations
SET revoked_at = now()
WHERE id = $1
  AND revoked_at IS NULL;
