CREATE TABLE "group_roles"
(
    "group_id"   CHAR(36)    NOT NULL,
    "role_id"    CHAR(36)    NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("group_id", "role_id"),
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE
);

-- Create trigger for updated_at
CREATE TRIGGER update_group_roles_updated_at
    BEFORE UPDATE
    ON "group_roles"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
