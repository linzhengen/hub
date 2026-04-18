CREATE TABLE "resources"
(
    "id"            CHAR(36) PRIMARY KEY,
    "parent_id"     CHAR(36)     NOT NULL DEFAULT '',
    "name"          VARCHAR(255) NOT NULL,
    "identifier"    VARCHAR(255) NOT NULL UNIQUE,
    "type"          VARCHAR(50)  NOT NULL,
    "path"          VARCHAR(255),
    "component"     VARCHAR(255),
    "display_order" INT,
    "description"   TEXT         NOT NULL DEFAULT '',
    "metadata"      JSONB,
    "status"        VARCHAR(50)  NOT NULL,
    "created_at"    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_resources_parent_id" ON "resources" ("parent_id");
CREATE INDEX "idx_resources_name" ON "resources" ("name");
CREATE INDEX "idx_resources_type" ON "resources" ("type");
CREATE INDEX "idx_resources_display_order" ON "resources" ("display_order");

-- Create trigger for updated_at
CREATE TRIGGER update_resources_updated_at
    BEFORE UPDATE
    ON "resources"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
