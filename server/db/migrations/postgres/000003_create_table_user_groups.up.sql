CREATE TABLE "user_groups"
(
    "user_id"    CHAR(36)    NOT NULL,
    "group_id"   CHAR(36)    NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("user_id", "group_id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE
);

-- Create trigger for updated_at
CREATE TRIGGER update_user_groups_updated_at
    BEFORE UPDATE
    ON "user_groups"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
