-- An access request is somebody asking to be put in a group, and somebody else
-- agreeing.
--
-- Before this table the only way into a group was an administrator doing it,
-- which records what happened but never why. The request carries the reason,
-- the term asked for and who agreed, so the audit entry for the grant has
-- something to point at.
--
-- This table is not read by an authorization decision - approving one writes to
-- `user_groups`, and that is the table with the `rbac_revisions` trigger - so it
-- deliberately has no trigger of its own.
CREATE TABLE "access_requests"
(
    "id"                 UUID PRIMARY KEY,
    -- Who asked. Not necessarily who the access is for: a manager asks on
    -- behalf of a report, and the assistant asks on behalf of whoever it is
    -- answering.
    "requester_user_id"  UUID         NOT NULL,
    -- Who the access is for.
    "subject_user_id"    UUID         NOT NULL,
    "group_id"           UUID         NOT NULL,
    -- Why. Required, because a request nobody explained is a request nobody can
    -- judge.
    "reason"             TEXT         NOT NULL,
    -- How long the access is wanted for. NULL asks for it permanently, which is
    -- a bigger ask and reads as one on the approval screen.
    "requested_until"    TIMESTAMPTZ,
    "status"             VARCHAR(50)  NOT NULL,
    -- Which surface the request came from. AI_CHAT is the one that matters:
    -- an approver has to know that a request was raised by the assistant from a
    -- conversation, because that is the case where the text that prompted it
    -- was written by somebody else.
    "origin"             VARCHAR(50)  NOT NULL,
    "session_id"         UUID,
    "decided_by_user_id" UUID,
    "decided_at"         TIMESTAMPTZ,
    "decision_comment"   TEXT         NOT NULL DEFAULT '',
    "created_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("requester_user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("subject_user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE
);

-- The queue an approver opens is "everything still pending", and the list a
-- requester opens is "mine, newest first".
CREATE INDEX "idx_access_requests_status" ON "access_requests" ("status", "created_at" DESC);
CREATE INDEX "idx_access_requests_requester" ON "access_requests" ("requester_user_id", "created_at" DESC);
CREATE INDEX "idx_access_requests_subject" ON "access_requests" ("subject_user_id", "created_at" DESC);

CREATE TRIGGER update_access_requests_updated_at
    BEFORE UPDATE
    ON "access_requests"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
