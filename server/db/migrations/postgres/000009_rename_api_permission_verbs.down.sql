-- Reverse 000009.
--
-- The rename is reversible; the rest is not, exactly. Rows created for verbs
-- that had no predecessor (RemoveRolesFromGroup) are dropped, and the singular
-- verbs deleted by the up migration cannot be brought back because nothing
-- recorded which roles held them. Rolling back therefore restores the old verb
-- names but not those grants; re-run `cli resource-import` afterwards and
-- reassign them.

-- 1. Undo the one-to-one renames.
UPDATE "permissions" p
SET "verb" = v.old_verb
FROM (
    VALUES ('AssignGroup', 'AddGroupsToUser'),
           ('UnassignGroup', 'RemoveGroupsFromUser'),
           ('AssignRolesToGroup', 'AddRolesToGroup')
) AS v(old_verb, new_verb)
WHERE p."verb" = v.new_verb
  AND EXISTS (SELECT 1 FROM "resources" r WHERE r."id" = p."resource_id" AND r."type" = 'api')
  AND NOT EXISTS (
      SELECT 1 FROM "permissions" existing
      WHERE existing."resource_id" = p."resource_id" AND existing."verb" = v.old_verb
  );

-- 2. Drop the verb that did not exist before this migration.
DELETE FROM "permissions" p
WHERE p."verb" = 'RemoveRolesFromGroup'
  AND EXISTS (SELECT 1 FROM "resources" r WHERE r."id" = p."resource_id" AND r."type" = 'api');
