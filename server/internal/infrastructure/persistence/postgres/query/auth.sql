-- name: SelectUserAuthorizedPolices :many
SELECT u.id, res.identifier, p.verb
FROM users AS u
         INNER JOIN user_groups AS ug ON u.id = ug.user_id
         INNER JOIN "groups" AS g ON g.id = ug.group_id AND g.status = 'Active'
         INNER JOIN group_roles AS gr ON ug.group_id = gr.group_id
         INNER JOIN role_permissions AS rp ON gr.role_id = rp.role_id
         INNER JOIN permissions AS p ON rp.permission_id = p.id
         INNER JOIN resources AS res ON p.resource_id = res.id AND res.status = 'Active'
WHERE u.id = $1
  AND u.status = 'Active';
