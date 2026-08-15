CREATE TABLE "user_groups"
(
    "user_id"    UUID        NOT NULL,
    "group_id"   UUID        NOT NULL,
    "expires_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("user_id", "group_id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE
);

-- expires_at ends the grant without anybody editing the graph. NULL means it
-- does not end.
--
-- It is deliberately not filtered by the authorization query. Expiry is the one
-- change to the graph that nobody writes, so the triggers on this table never
-- fire for it and `rbac_revisions` never moves. A `expires_at > now()` predicate
-- in SQL would therefore be memoised by the policy cache and keep serving a
-- lapsed grant for the length of the TTL - exactly the failure `rbac_revisions`
-- exists to prevent. The rows are returned with their expiry attached and
-- `auth.Service.Enforce` drops them at the moment it decides, so a decision is
-- always made against the clock rather than against whenever the cache was
-- filled.

-- Sweeping lapsed rows is a separate question from enforcing them, and this
-- index is what a sweeper - or a console listing what is about to lapse - would
-- read. Partial, because a row with no expiry is never the answer to "what
-- expires soon".
CREATE INDEX "idx_user_groups_expires_at" ON "user_groups" ("expires_at")
    WHERE "expires_at" IS NOT NULL;

-- Create trigger for updated_at
CREATE TRIGGER update_user_groups_updated_at
    BEFORE UPDATE
    ON "user_groups"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
