-- Realign the API permission verbs with the renamed rpcs.
--
-- The verb of an API permission is the rpc name, so renaming an rpc orphans
-- every permission that named it: the row survives but nothing enforces
-- against it any more, and the roles holding it silently lose the ability.
-- These statements carry the grants across instead.
--
-- Only permissions on API resources are touched. Menu resources use the verb
-- "view" and are unaffected, as is the wildcard "*" that the admin role holds.

-- 1. Verbs with a one-to-one successor. The permission row keeps its id, so
--    every role_permissions grant follows it automatically.
UPDATE "permissions" p
SET "verb" = v.new_verb
FROM (
    VALUES ('AssignGroup', 'AddGroupsToUser'),
           ('UnassignGroup', 'RemoveGroupsFromUser'),
           ('AssignRolesToGroup', 'AddRolesToGroup')
) AS v(old_verb, new_verb)
WHERE p."verb" = v.old_verb
  AND EXISTS (SELECT 1 FROM "resources" r WHERE r."id" = p."resource_id" AND r."type" = 'api')
  -- Skip when the successor already exists on this resource, so the rename
  -- cannot produce two rows for the same verb.
  AND NOT EXISTS (
      SELECT 1 FROM "permissions" existing
      WHERE existing."resource_id" = p."resource_id" AND existing."verb" = v.new_verb
  );

-- 2. The singular rpcs are gone; their plural counterparts do the same job.
--    Give the plural verb to any role that only held the singular one, so that
--    nobody loses an ability they had.
INSERT INTO "permissions" ("id", "verb", "resource_id", "description")
SELECT gen_random_uuid()::text, v.new_verb, p."resource_id", v.description
FROM "permissions" p
         JOIN (
    VALUES ('AssignRole', 'AddRolesToGroup', 'Grant roles to a group, keeping the roles it already has.'),
           ('AssignPermission', 'AddPermissionsToRole', 'Grant permissions to a role, keeping the ones it already has.')
) AS v(old_verb, new_verb, description) ON p."verb" = v.old_verb
WHERE EXISTS (SELECT 1 FROM "resources" r WHERE r."id" = p."resource_id" AND r."type" = 'api')
  AND NOT EXISTS (
      SELECT 1 FROM "permissions" existing
      WHERE existing."resource_id" = p."resource_id" AND existing."verb" = v.new_verb
  )
GROUP BY v.new_verb, p."resource_id", v.description;

INSERT INTO "role_permissions" ("role_id", "permission_id")
SELECT DISTINCT rp."role_id", successor."id"
FROM "role_permissions" rp
         JOIN "permissions" p ON p."id" = rp."permission_id"
         JOIN (
    VALUES ('AssignRole', 'AddRolesToGroup'),
           ('AssignPermission', 'AddPermissionsToRole')
) AS v(old_verb, new_verb) ON p."verb" = v.old_verb
         JOIN "permissions" successor
              ON successor."resource_id" = p."resource_id" AND successor."verb" = v.new_verb
ON CONFLICT DO NOTHING;

-- 3. RemoveRolesFromGroup is new. Revoking a role from a group used to be done
--    by replacing the group's whole role set through AssignRolesToGroup, which
--    is now AddRolesToGroup - so whoever can add roles could already remove
--    them, and must keep being able to.
INSERT INTO "permissions" ("id", "verb", "resource_id", "description")
SELECT gen_random_uuid()::text, 'RemoveRolesFromGroup', p."resource_id", 'Revoke roles from a group.'
FROM "permissions" p
WHERE p."verb" = 'AddRolesToGroup'
  AND NOT EXISTS (
      SELECT 1 FROM "permissions" existing
      WHERE existing."resource_id" = p."resource_id" AND existing."verb" = 'RemoveRolesFromGroup'
  );

INSERT INTO "role_permissions" ("role_id", "permission_id")
SELECT DISTINCT rp."role_id", successor."id"
FROM "role_permissions" rp
         JOIN "permissions" p ON p."id" = rp."permission_id" AND p."verb" = 'AddRolesToGroup'
         JOIN "permissions" successor
              ON successor."resource_id" = p."resource_id" AND successor."verb" = 'RemoveRolesFromGroup'
ON CONFLICT DO NOTHING;

-- 4. Drop the verbs that no longer name an rpc. role_permissions cascades.
DELETE FROM "permissions" p
WHERE p."verb" IN ('AssignRole', 'AssignPermission')
  AND EXISTS (SELECT 1 FROM "resources" r WHERE r."id" = p."resource_id" AND r."type" = 'api');
