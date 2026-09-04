-- name: CreateAuditLog :one
INSERT INTO audit_logs (actor_user_id,
                        org_id,
                        channel,
                        ai_session_id,
                        resource,
                        action,
                        target_id,
                        arguments,
                        succeeded,
                        error,
                        client,
                        approval_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;
