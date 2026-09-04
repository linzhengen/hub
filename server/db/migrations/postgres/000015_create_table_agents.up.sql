-- An agent is a program that calls the API: an A2A agent hub deploys, a
-- sub-agent one of those calls, an assistant wired to a set of tools.
--
-- **It acts as itself, not on anybody's behalf.** An agent authenticates as its
-- own Keycloak client and is authorized against its own grants, exactly as a
-- service account is. There is no way for it to say "I am acting for this
-- person", because the row that would make such a claim checkable does not
-- exist yet: `delegations` arrives in a later migration, and until it does, an
-- audit record of an agent's action names the agent and stops there.
--
-- What an investigation *can* follow today is who is answerable for the agent:
-- `audit_logs.actor_user_id` joins `agents.user_id`, and the row carries
-- `created_by_user_id`. What it cannot answer is which person caused this
-- particular call - every invocation of one agent collapses to one actor.
--
-- That is tolerable only while nothing invokes an agent for somebody. Anything
-- that does - a deployed runtime, a console that runs one, an A2A endpoint -
-- has to wait for the delegation chain, or it builds a confused deputy: a
-- caller wielding the agent's permissions without holding them, and without a
-- trace.
--
-- This table is a register, not an authentication mechanism. It is the same
-- shape `service_accounts` (`000014`) established, and for the same reason:
-- authentication stays Keycloak's, and the `user_id` column is what makes the
-- rest of hub work unchanged. An agent with a `users` row joins groups, holds
-- roles, receives time-bounded grants, can be the subject of an access request
-- and appears in the audit log as an actor - through exactly the machinery a
-- person does.
--
-- **Do not add a new kind of actor.** The moment there is one, human
-- authorization and machine authorization fork, and a hole opened in one stops
-- being visible from the other.
--
-- Two columns are new relative to `service_accounts`:
--
--   `org_id`          an agent belongs to a tenant. `service_accounts`
--                     predates organizations and is implicitly the platform's;
--                     an agent is not, so the boundary is a column.
--   `parent_agent_id` a sub-agent records the agent that owns it. Nothing
--                     enforces anything on it yet - the delegation chain that
--                     bounds a sub-agent's authority arrives with
--                     `delegations`. It is here now because adding a
--                     self-reference later means rewriting the rows that
--                     should have had one.
--
-- There is deliberately no `rbac_revisions` trigger. An authorization decision
-- does not read this table: it reads the agent's `users` row and the graph
-- above it, all of which already have triggers. A trigger here would invalidate
-- every cached policy in the installation whenever an agent was renamed.
CREATE TABLE "agents"
(
    "id"                 UUID PRIMARY KEY,
    -- The tenant the agent belongs to. RESTRICT rather than CASCADE: dropping
    -- an organization must not silently take the register away and leave the
    -- Keycloak clients working, which is the one failure the whole delete path
    -- is ordered to avoid.
    "org_id"             UUID         NOT NULL,
    -- The hub user the agent acts as: the Keycloak client's own service account
    -- user, stored in `users`.
    "user_id"            UUID         NOT NULL UNIQUE,
    -- Unique across the installation rather than within the organization: the
    -- name becomes a Keycloak client id, and Keycloak's namespace is the realm.
    "name"               VARCHAR(255) NOT NULL UNIQUE,
    -- What the agent is for. It is also what the public Agent Card will say, so
    -- it is a description meant to be read rather than an internal note.
    "description"        TEXT         NOT NULL DEFAULT '',
    -- The Keycloak client this authenticates as. client_id is what an operator
    -- configures; keycloak_id is Keycloak's internal handle, kept because every
    -- admin call needs it and deriving it again means another round trip.
    "client_id"          VARCHAR(255) NOT NULL UNIQUE,
    "keycloak_id"        VARCHAR(255) NOT NULL,
    -- How the agent proves it is the client.
    --
    -- Only CLIENT_SECRET is issued today. The column exists because a shared
    -- secret is the part of this design that most clearly falls short of what
    -- is now recommended for machine identities - short-lived, non-shared,
    -- key-bound credentials - and PRIVATE_KEY_JWT is the migration Keycloak
    -- already supports. Recording the method per agent is what lets the two
    -- coexist while agents are moved across, instead of a flag day.
    "auth_method"        VARCHAR(32)  NOT NULL DEFAULT 'CLIENT_SECRET',
    -- When the current secret was issued. hub never stores the secret itself,
    -- so this is the only thing it can say about its age - and an ageing
    -- credential nobody can see is one nobody rotates.
    --
    -- There is no `last_authenticated_at` beside it. Writing one would mean a
    -- database write on every authenticated request, and `audit_logs` already
    -- records the agent as an actor whenever it changes anything.
    "secret_rotated_at"  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- The agent that owns this one, NULL for a root agent. RESTRICT for the
    -- same reason as org_id: cascading would delete the child's row while its
    -- Keycloak client kept working.
    "parent_agent_id"    UUID,
    -- Who set it up. An agent nobody owns is the one nobody turns off, so the
    -- reference is RESTRICT: a person cannot be removed while agents still name
    -- them as their controller. Reassigning them is the offboarding step that
    -- keeps every agent attached to a human.
    "created_by_user_id" UUID         NOT NULL,
    "created_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON DELETE RESTRICT,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("parent_agent_id") REFERENCES "agents" ("id") ON DELETE RESTRICT,
    FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("id") ON DELETE RESTRICT,
    CONSTRAINT "agents_auth_method_check"
        CHECK ("auth_method" IN ('CLIENT_SECRET', 'PRIVATE_KEY_JWT'))
);

CREATE INDEX "idx_agents_org_id" ON "agents" ("org_id");
CREATE INDEX "idx_agents_parent_agent_id" ON "agents" ("parent_agent_id");
CREATE INDEX "idx_agents_created_by" ON "agents" ("created_by_user_id");

CREATE TRIGGER update_agents_updated_at
    BEFORE UPDATE
    ON "agents"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
