CREATE TABLE "users"
(
    "id"         CHAR(36) PRIMARY KEY,
    "username"   VARCHAR(255) NOT NULL,
    "email"      VARCHAR(255) NOT NULL UNIQUE,
    "status"     VARCHAR(50)  NOT NULL,
    "created_at" TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_users_username" ON "users" ("username");

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON "users"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
