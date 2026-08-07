-- A counter that changes whenever anything the authorization decision reads
-- changes, so a cache can tell in one cheap query whether what it holds is
-- still current.
--
-- Enforcing a permission means a seven-way join over users, groups, roles,
-- permissions and resources, once per request. Caching the result needs a way
-- to notice a grant or a revocation, and time alone is a poor one: a short TTL
-- throws the cache away constantly, a long one leaves a revoked user working.
CREATE TABLE "rbac_revisions"
(
    -- One row, forever. The primary key exists to say so.
    "id"         SMALLINT PRIMARY KEY DEFAULT 1,
    "revision"   BIGINT      NOT NULL DEFAULT 1,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "rbac_revisions_single_row" CHECK ("id" = 1)
);

INSERT INTO "rbac_revisions" ("id") VALUES (1);

-- The trigger lives in the database rather than in the use cases because the
-- graph is edited from more places than the API: `cli seed`, `resource-import`,
-- migrations, and whatever an operator runs by hand. Any of those forgetting to
-- bump the counter would leave a stale cache serving permissions that no longer
-- exist, which is the failure this table is meant to prevent.
CREATE OR REPLACE FUNCTION bump_rbac_revision() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE "rbac_revisions"
    SET "revision"   = "revision" + 1,
        "updated_at" = CURRENT_TIMESTAMP
    WHERE "id" = 1;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Statement-level, not row-level: bulk edits should cost one bump, not one per
-- row. The counter only has to change, not to count anything.
--
-- Every bump updates the same row, so RBAC writes serialise against each other.
-- These are administrative operations, so that is a fair price for never
-- serving a revoked permission.
--
-- One statement may bump more than once: deleting a permission cascades to
-- role_permissions, and a statement trigger fires even when the cascaded
-- statement matches no rows. That is fine. Readers only ask whether the number
-- changed, never by how much.
CREATE TRIGGER bump_rbac_revision_on_users
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "users"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_user_groups
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "user_groups"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_groups
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "groups"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_group_roles
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "group_roles"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_roles
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "roles"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_role_permissions
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "role_permissions"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_permissions
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "permissions"
EXECUTE FUNCTION bump_rbac_revision();

CREATE TRIGGER bump_rbac_revision_on_resources
    AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
    ON "resources"
EXECUTE FUNCTION bump_rbac_revision();
