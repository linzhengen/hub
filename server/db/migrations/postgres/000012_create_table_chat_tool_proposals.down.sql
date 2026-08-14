ALTER TABLE chat_sessions DROP COLUMN IF EXISTS tokens_used;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS approval_id;
DROP TABLE IF EXISTS chat_tool_proposals;
