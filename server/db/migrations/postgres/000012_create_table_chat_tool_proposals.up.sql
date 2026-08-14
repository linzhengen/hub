-- chat_tool_proposals holds a change the assistant wants to make and has not
-- made. The row exists because grpc-gateway's server-streaming is one way: the
-- answer stream has to end for the user to reply to it, so the state of the
-- paused conversation has to outlive the request that produced it.
--
-- continuation is opaque to everything but the LLM backend that wrote it: what
-- that backend needs to take the conversation back up. Keeping it opaque is
-- what stops one provider's message format from becoming part of the schema.
CREATE TABLE chat_tool_proposals (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id   UUID        NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    tool_name    TEXT        NOT NULL,
    arguments    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    continuation BYTEA       NOT NULL,
    -- A proposal is decided once. Re-answering a decided one must not run the
    -- change a second time, which is why the status is stored rather than
    -- inferred from whether a decision arrived.
    status       TEXT        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at   TIMESTAMPTZ
);

CREATE INDEX idx_chat_tool_proposals_session_id ON chat_tool_proposals (session_id, created_at DESC);

-- approval_id ties a recorded change to the proposal the user approved, so the
-- audit log answers "did a person agree to this" and not only "it came through
-- the assistant". Null for every change that did not go through the flow.
ALTER TABLE audit_logs ADD COLUMN approval_id UUID;

-- tokens_used is what the conversation has spent. It bounds a session that
-- keeps re-reading a large directory, honestly or otherwise, and it is per
-- session so that hitting the cap costs one conversation rather than the
-- assistant.
ALTER TABLE chat_sessions ADD COLUMN tokens_used BIGINT NOT NULL DEFAULT 0;
