-- A delegation is the fact that an agent may act on one person's behalf.
--
-- It is a row rather than a token, and that is the whole design. A row can be
-- revoked now: `rbac_revisions` bumps on the write and every cache drops what
-- it holds within a second, so there is no residual lifetime to reason about. A
-- token would have one, and the only ways to shorten it are to make it short
-- (and refresh constantly) or to keep a revocation list - which is a table
-- again, only one the authorization path has to consult in addition to the
-- token rather than instead of it.
--
-- It also means authentication is unchanged. An agent authenticates as itself,
-- through the client credentials grant, exactly as `000015` left it. What it
-- gains is the ability to *say* it is acting for someone, and the reason it
-- cannot say that about an arbitrary person is that the row does not exist.
--
-- The effective authority of an agent acting under a delegation is
--
--     the agent's own grants  ∩  the principal's grants  ∩  this delegation
--
-- and the intersection is taken **when the decision is made**, never when the
-- row is written. Folding it in at write time would keep the agent working
-- after the principal lost the access it was derived from - the same trap that
-- `user_groups.expires_at` avoids by not being filtered in SQL.
--
-- The third term is the reason `delegation_permissions` exists below.
CREATE TABLE "delegations"
(
    "id"                 UUID        PRIMARY KEY,
    -- The agent that may act. Removing the agent removes its delegations:
    -- authority with no holder is not worth keeping, and dropping it fails
    -- closed.
    "agent_id"           UUID        NOT NULL,
    -- Whose authority is being lent. Only this person can create the row
    -- through the API, so an agent cannot be handed a stranger's authority by
    -- a third party who happens to be an administrator.
    "principal_user_id"  UUID        NOT NULL,
    -- Who wrote the row. Today it is always the principal; it is a separate
    -- column because the other route into this table - an approved access
    -- request - has somebody else settle it, and "who agreed to this" is the
    -- first question asked when a delegation looks wrong.
    "granted_by_user_id" UUID        NOT NULL,
    -- The organization the delegation is answerable in, copied from the agent.
    -- It is denormalised for the same reason `audit_logs.org_id` is: the
    -- listing is filtered by it on every read, and an agent's organization
    -- cannot change (there is no rpc that moves one).
    "org_id"             UUID        NOT NULL,
    -- Why this exists, in the words of the person who agreed to it. A
    -- delegation nobody can explain is one nobody dares revoke.
    "reason"             TEXT        NOT NULL,
    -- How many agents may stand in the chain. 1 is the agent named here and no
    -- further; each step up admits one more sub-agent between it and the
    -- principal. The bound is stored per delegation because it is the
    -- principal's decision, not a global setting. Nothing reads it yet - the
    -- chain arrives with the authorization change.
    "max_depth"          SMALLINT    NOT NULL DEFAULT 1,
    -- When the delegation stops, NULL for one that does not. The meaning is
    -- `user_groups.expires_at`'s exactly, and like it, it is **not** filtered
    -- in SQL: the decision drops a lapsed row against the clock, because an
    -- expiry is the one change nobody writes and so the revision counter never
    -- moves for it.
    --
    -- The API does not offer NULL. An endless delegation is the exception, so
    -- it is not reachable from the ordinary path even though the column admits
    -- it.
    "expires_at"         TIMESTAMPTZ,
    -- When it was revoked, NULL while live. Revocation is a write, so unlike
    -- the expiry it does bump the revision counter, and a decision may
    -- therefore filter on it in SQL.
    --
    -- The row is kept rather than deleted: "this agent could act for me until
    -- Tuesday" is exactly what an investigation needs afterwards. Who revoked
    -- it is not a column here - `audit_logs` records that, and recording it
    -- twice is two places to disagree.
    "revoked_at"         TIMESTAMPTZ,
    "created_at"         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("agent_id") REFERENCES "agents" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("principal_user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("granted_by_user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON DELETE CASCADE,
    CONSTRAINT "delegations_max_depth_positive" CHECK ("max_depth" >= 1)
);

CREATE INDEX "idx_delegations_agent_id" ON "delegations" ("agent_id");
CREATE INDEX "idx_delegations_principal_user_id" ON "delegations" ("principal_user_id");
CREATE INDEX "idx_delegations_org_id" ON "delegations" ("org_id");

CREATE TRIGGER update_delegations_updated_at
    BEFORE UPDATE
    ON "delegations"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- What the delegation actually carries: a subset of the permissions the
-- principal holds.
--
-- It points at `permissions` rather than storing patterns of its own because
-- those rows *are* the vocabulary a decision is made in - a permission is a
-- resource pattern and a verb, which is exactly what the enforcement matches
-- against. Inventing a second way to write the same thing would mean a second
-- matcher, and two matchers eventually disagree.
--
-- **There is no "everything" delegation.** A row with no permissions grants
-- nothing, and the API refuses to create one. The alternative - treating the
-- empty set as the principal's full authority - fails open: a bug that skipped
-- these inserts would silently produce the widest delegation in the system
-- instead of the narrowest. This is the table whose entire purpose is to bound
-- authority, so it fails closed.
CREATE TABLE "delegation_permissions"
(
    "delegation_id" UUID        NOT NULL,
    "permission_id" UUID        NOT NULL,
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("delegation_id", "permission_id"),
    FOREIGN KEY ("delegation_id") REFERENCES "delegations" ("id") ON DELETE CASCADE,
    -- Removing a permission narrows every delegation that named it, which is
    -- the safe direction for a cascade to run in.
    FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE
);

CREATE TRIGGER update_delegation_permissions_updated_at
    BEFORE UPDATE
    ON "delegation_permissions"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Both tables are read by an authorization decision, so both need the counter.
-- Without these, revoking a delegation would keep working for the length of the
-- policy cache TTL - which is the failure `rbac_revisions` exists to prevent,
-- and the one that matters most here: revocation is the whole reason a
-- delegation is a row rather than a token.
CREATE TRIGGER bump_rbac_revision_on_delegations
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "delegations"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_delegation_permissions
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "delegation_permissions"
EXECUTE FUNCTION bump_rbac_revision();
