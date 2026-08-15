-- A service account is a machine that calls the API: CI, a job, another
-- service.
--
-- Before this table the only way to give a machine credentials was to create a
-- Keycloak client by hand. That works - the CLI has spoken client_credentials
-- all along - but nothing in hub knew the client existed, so nobody could ask
-- what it was for, who set it up, or what it had done.
--
-- Authentication is still Keycloak's. This table is the register: it says which
-- Keycloak client a machine authenticates as, who created it and why, and it
-- carries the `user_id` that makes the rest of hub work unchanged.
--
-- That last column is the important one. A Keycloak client with service
-- accounts enabled has a user of its own, and hub stores that user in `users`
-- like any other. So a service account joins groups, holds roles and appears in
-- the audit log through exactly the machinery a person does - no second
-- authorization path, and no second kind of actor.
CREATE TABLE "service_accounts"
(
    "id"                 UUID PRIMARY KEY,
    -- The hub user the machine acts as: the Keycloak client's own service
    -- account user, stored in `users`.
    "user_id"            UUID         NOT NULL UNIQUE,
    "name"               VARCHAR(255) NOT NULL UNIQUE,
    "description"        TEXT         NOT NULL DEFAULT '',
    -- The Keycloak client this authenticates as. client_id is what an operator
    -- puts in HUB_OIDC_CLIENT_ID; keycloak_id is Keycloak's internal handle,
    -- kept because every admin call needs it and deriving it again means
    -- another round trip.
    "client_id"          VARCHAR(255) NOT NULL UNIQUE,
    "keycloak_id"        VARCHAR(255) NOT NULL,
    -- Who set it up. A machine nobody owns is the one nobody turns off.
    "created_by_user_id" UUID         NOT NULL,
    "created_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE INDEX "idx_service_accounts_created_by" ON "service_accounts" ("created_by_user_id");

CREATE TRIGGER update_service_accounts_updated_at
    BEFORE UPDATE
    ON "service_accounts"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
