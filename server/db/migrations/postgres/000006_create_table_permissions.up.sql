CREATE TABLE "permissions"
(
    "id"          CHAR(36) PRIMARY KEY,
    "verb"        VARCHAR(255) NOT NULL,
    "resource_id" CHAR(36)     NOT NULL,
    "description" TEXT         NOT NULL DEFAULT '',
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("resource_id") REFERENCES "resources" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_permissions_verb" ON "permissions" ("verb");
CREATE INDEX "idx_permissions_resource_id" ON "permissions" ("resource_id");

-- Create trigger for updated_at
CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE
    ON "permissions"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
