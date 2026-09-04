-- An organization is the boundary a permission is held within, and a group is
-- the edge that joins a user to one.
--
-- They are created together because neither is meaningful without the other: a
-- group belongs to exactly one organization, so every route from a user to a
-- permission passes through one, and an authorization decision can be about a
-- place rather than only about a verb. Splitting them would mean a group that
-- briefly has no boundary.
--
-- `kind` is the only column that distinguishes a business tenant from an
-- individual. PERSONAL is a one-person organization, which is what lets an
-- individual user travel the same authorization path a company's staff do
-- instead of needing a second one. There is deliberately no branch in the code
-- on this column; it exists to be shown to a human and to be reported on.
--
-- PLATFORM is the one that carries a rule: a grant reached through it is
-- answerable about every organization. There is one, it is inserted below, and
-- the seed's groups belong to it, so an operator's grants reach every tenant.
CREATE TABLE "organizations"
(
    "id"          UUID PRIMARY KEY,
    "name"        VARCHAR(255) NOT NULL,
    -- The handle an operator types and a URL carries. Unique because it names
    -- the tenant in places a UUID would be unusable.
    "slug"        VARCHAR(255) NOT NULL UNIQUE,
    "kind"        VARCHAR(50)  NOT NULL CHECK ("kind" IN ('PLATFORM', 'BUSINESS', 'PERSONAL')),
    "description" TEXT         NOT NULL DEFAULT '',
    "status"      VARCHAR(50)  NOT NULL,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_organizations_name" ON "organizations" ("name");
CREATE INDEX "idx_organizations_kind" ON "organizations" ("kind");

-- The platform organization is created here rather than in the seed, because
-- `groups.org_id` is NOT NULL and the seed's groups have to land somewhere. A
-- schema whose first insert depends on `cli seed` having run is a schema that
-- cannot be migrated on its own.
INSERT INTO "organizations" ("id", "name", "slug", "kind", "description", "status")
VALUES ('00000000-0000-0000-0000-000000000001',
        'platform',
        'platform',
        'PLATFORM',
        'The operator of this installation. Its grants apply in every organization.',
        'Active');

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE
    ON "organizations"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE "groups"
(
    "id"          UUID PRIMARY KEY,
    "name"        VARCHAR(255) NOT NULL,
    "description" TEXT         NOT NULL DEFAULT '',
    "status"      VARCHAR(50)  NOT NULL,
    -- The organization this group belongs to.
    --
    -- NOT NULL, because a group with no organization would be a route to a
    -- permission that no boundary contains. It is set once: moving a group
    -- between organizations would carry every member's access across a tenant
    -- boundary in a single statement, which is a migration rather than an edit,
    -- so `UpdateGroup` does not touch it.
    "org_id"      UUID         NOT NULL,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Offboarding a tenant takes its groups with it, and the memberships and
    -- role grants below them already cascade from `groups`.
    FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_groups_name" ON "groups" ("name");
CREATE INDEX "idx_groups_org_id" ON "groups" ("org_id");

-- Create trigger for updated_at
CREATE TRIGGER update_groups_updated_at
    BEFORE UPDATE
    ON "groups"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
