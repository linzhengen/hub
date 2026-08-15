-- name: SelectUserAuthorizedPolices :many
-- The expiry is not filtered here on purpose: see the note on expires_at in
-- the user_groups and group_roles migrations.
--
-- Both edges of the route are returned rather than LEAST() of them, so that
-- "the route expires with whichever edge expires first" is stated once, in Go,
-- where it is under test - and so that a NULL keeps meaning "never" rather than
-- depending on how LEAST treats it.
SELECT u.id, res.identifier, p.verb, ug.expires_at AS group_expires_at, gr.expires_at AS role_expires_at
FROM users AS u
         INNER JOIN user_groups AS ug ON u.id = ug.user_id
         INNER JOIN "groups" AS g ON g.id = ug.group_id AND g.status = 'Active'
         INNER JOIN group_roles AS gr ON ug.group_id = gr.group_id
         INNER JOIN role_permissions AS rp ON gr.role_id = rp.role_id
         INNER JOIN permissions AS p ON rp.permission_id = p.id
         INNER JOIN resources AS res ON p.resource_id = res.id AND res.status = 'Active'
WHERE u.id = $1
  AND u.status = 'Active';

-- name: SelectUserAccessPaths :many
SELECT g.id     AS group_id,
       g.name   AS group_name,
       r.id     AS role_id,
       r.name   AS role_name,
       p.id     AS permission_id,
       res.identifier,
       p.verb,
       ug.expires_at AS group_expires_at,
       gr.expires_at AS role_expires_at
FROM users AS u
         INNER JOIN user_groups AS ug ON u.id = ug.user_id
         INNER JOIN "groups" AS g ON g.id = ug.group_id AND g.status = 'Active'
         INNER JOIN group_roles AS gr ON ug.group_id = gr.group_id
         INNER JOIN roles AS r ON r.id = gr.role_id
         INNER JOIN role_permissions AS rp ON gr.role_id = rp.role_id
         INNER JOIN permissions AS p ON rp.permission_id = p.id
         INNER JOIN resources AS res ON p.resource_id = res.id AND res.status = 'Active'
WHERE u.id = $1
  AND u.status = 'Active'
ORDER BY g.name, r.name, res.identifier, p.verb;

-- name: SelectAccessPaths :many
SELECT g.id     AS group_id,
       g.name   AS group_name,
       r.id     AS role_id,
       r.name   AS role_name,
       p.id     AS permission_id,
       res.identifier,
       p.verb,
       gr.expires_at
FROM "groups" AS g
         INNER JOIN group_roles AS gr ON g.id = gr.group_id
         INNER JOIN roles AS r ON r.id = gr.role_id
         INNER JOIN role_permissions AS rp ON gr.role_id = rp.role_id
         INNER JOIN permissions AS p ON rp.permission_id = p.id
         INNER JOIN resources AS res ON p.resource_id = res.id AND res.status = 'Active'
WHERE g.status = 'Active'
ORDER BY g.name, r.name, res.identifier, p.verb;

-- name: SelectMemberships :many
SELECT u.id,
       u.username,
       ug.group_id,
       ug.expires_at
FROM users AS u
         INNER JOIN user_groups AS ug ON u.id = ug.user_id
WHERE u.status = 'Active'
ORDER BY u.username, u.id;
