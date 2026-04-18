CREATE TABLE "groups"
(
    "id"          CHAR(36) PRIMARY KEY,
    "name"        VARCHAR(255) NOT NULL,
    "description" TEXT         NOT NULL DEFAULT '',
    "status"      VARCHAR(50)  NOT NULL,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_groups_name" ON "groups" ("name");

-- Create trigger for updated_at
CREATE TRIGGER update_groups_updated_at
    BEFORE UPDATE
    ON "groups"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
