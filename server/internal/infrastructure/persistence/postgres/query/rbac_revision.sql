-- name: SelectRbacRevision :one
SELECT revision
FROM rbac_revisions
WHERE id = 1;
