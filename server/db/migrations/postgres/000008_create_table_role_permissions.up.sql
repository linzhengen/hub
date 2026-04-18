CREATE TABLE "role_permissions"
(
    "role_id"       CHAR(36)    NOT NULL,
    "permission_id" CHAR(36)    NOT NULL,
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("role_id", "permission_id"),
    FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE
);

-- Create trigger for updated_at
CREATE TRIGGER update_role_permissions_updated_at
    BEFORE UPDATE
    ON "role_permissions"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
