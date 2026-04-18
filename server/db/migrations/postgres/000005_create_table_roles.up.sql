CREATE TABLE "roles"
(
    "id"          CHAR(36) PRIMARY KEY,
    "name"        VARCHAR(255) NOT NULL,
    "description" TEXT         NOT NULL DEFAULT '',
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_roles_name" ON "roles" ("name");

-- Create trigger for updated_at
CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE
    ON "roles"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
