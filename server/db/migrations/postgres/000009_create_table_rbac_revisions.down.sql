DROP TRIGGER IF EXISTS "bump_rbac_revision_on_resources" ON "resources";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_permissions" ON "permissions";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_role_permissions" ON "role_permissions";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_roles" ON "roles";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_group_roles" ON "group_roles";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_groups" ON "groups";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_organizations" ON "organizations";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_user_groups" ON "user_groups";
DROP TRIGGER IF EXISTS "bump_rbac_revision_on_users" ON "users";

DROP FUNCTION IF EXISTS bump_rbac_revision();

DROP TABLE IF EXISTS "rbac_revisions";
