-- audit_logs records every attempt to change the RBAC graph: who made it,
-- through which channel, what they asked for and whether it was allowed.
--
-- actor_user_id deliberately carries no foreign key. An audit record has to
-- outlive the account it describes - "who granted this, and are they still
-- here?" is exactly the question asked after someone leaves - and a cascade or
-- a restrict would either erase the record or block the deletion.
--
-- The table needs no rbac_revisions trigger: no authorization decision reads
-- it, so a write here cannot make a cached policy stale.
CREATE TABLE audit_logs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID        NOT NULL,
    channel       TEXT        NOT NULL CHECK (channel IN ('api', 'ai_chat')),
    -- The chat session the assistant was acting in, when channel is 'ai_chat'.
    -- It is what ties a change back to the conversation that asked for it.
    ai_session_id UUID,
    resource      TEXT        NOT NULL,
    action        TEXT        NOT NULL,
    target_id     TEXT        NOT NULL DEFAULT '',
    -- The request as it arrived, with secrets removed. Storing the arguments
    -- rather than a before/after diff keeps the record at the level the caller
    -- actually operated at, which is what has to be explained afterwards.
    arguments     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    -- A refused attempt is recorded too: an attempt to grant oneself admin that
    -- was denied is the event an audit log exists to show.
    succeeded     BOOLEAN     NOT NULL,
    error         TEXT        NOT NULL DEFAULT '',
    client        TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The log is read newest first, either whole or narrowed to one actor, one kind
-- of operation, or one affected entity.
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX idx_audit_logs_operation ON audit_logs (resource, action, created_at DESC);
CREATE INDEX idx_audit_logs_target ON audit_logs (target_id, created_at DESC);
