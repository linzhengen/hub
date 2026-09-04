CREATE TABLE "roles"
(
    "id"          UUID PRIMARY KEY,
    "name"        VARCHAR(255) NOT NULL,
    "description" TEXT         NOT NULL DEFAULT '',
    -- The organization that defines this role.
    --
    -- Nullable, unlike a group's: a group is where the tenant boundary lives so
    -- it always has one, whereas a role is a named bundle of permissions that
    -- can sensibly be shared. NULL is a role this installation provides to
    -- every organization - `admin-role` from the seed is the first - and a
    -- value is a role a tenant defined for itself. Both live in one table so
    -- that a group's grants are looked up the same way whichever it holds.
    "org_id"      UUID,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_roles_name" ON "roles" ("name");

-- Partial: a shared role is never the answer to "which roles does this tenant
-- define?", which is the only query this index serves.
CREATE INDEX "idx_roles_org_id" ON "roles" ("org_id")
    WHERE "org_id" IS NOT NULL;

-- Create trigger for updated_at
CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE
    ON "roles"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
