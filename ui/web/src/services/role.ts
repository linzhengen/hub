import { roleServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-role-v1-service';
import type { RequestParameters } from '@/api/helper';

// Re-export schema types for backward compatibility
export type Role = components['schemas']['v1Role'];
export type ListRolesResponse = components['schemas']['v1ListRoleResponse'];
export type GetRoleResponse = components['schemas']['v1GetRoleResponse'];
export type CreateRoleRequest = components['schemas']['v1CreateRoleRequest'];
export type CreateRoleResponse = components['schemas']['v1CreateRoleResponse'];
export type UpdateRoleRequest = components['schemas']['RoleServiceUpdateRoleBody'];
export type UpdateRoleResponse = components['schemas']['v1UpdateRoleResponse'];
export type AssignPermissionRequest = components['schemas']['RoleServiceAssignPermissionBody'];
export type AssignPermissionResponse = components['schemas']['v1AssignPermissionResponse'];
export type AddPermissionsToRoleRequest = components['schemas']['RoleServiceAddPermissionsToRoleBody'];
export type AddPermissionsToRoleResponse = components['schemas']['v1AddPermissionsToRoleResponse'];
export type RemovePermissionsFromRoleRequest = components['schemas']['RoleServiceRemovePermissionsFromRoleBody'];
export type RemovePermissionsFromRoleResponse = components['schemas']['v1RemovePermissionsFromRoleResponse'];
export type DeleteRoleResponse = components['schemas']['v1DeleteRoleResponse'];

// Helper type for list roles parameters (query)
export type ListRolesParams = RequestParameters<paths, '/api/v1/roles', 'get'>;

export const roleService = {
  listRoles: (params?: ListRolesParams) => apiRequest<ListRolesResponse>(ops.listRole, { query: params }),
  getRole: (id: string) => apiRequest<GetRoleResponse>(ops.getRole, { path: { id } }),
  createRole: (data: CreateRoleRequest) => apiRequest<CreateRoleResponse>(ops.createRole, { body: data }),
  updateRole: (id: string, data: UpdateRoleRequest) => apiRequest<UpdateRoleResponse>(ops.updateRole, { path: { id }, body: data }),
  deleteRole: (id: string) => apiRequest<DeleteRoleResponse>(ops.deleteRole, { path: { id } }),

  assignPermission: (id: string, data: AssignPermissionRequest) => apiRequest<AssignPermissionResponse>(ops.assignPermission, { path: { id }, body: data }),
  addPermissionsToRole: (id: string, data: AddPermissionsToRoleRequest) => apiRequest<AddPermissionsToRoleResponse>(ops.addPermissionsToRole, { path: { roleId: id }, body: data }),
  removePermissionsFromRole: (id: string, data: RemovePermissionsFromRoleRequest) => apiRequest<RemovePermissionsFromRoleResponse>(ops.removePermissionsFromRole, { path: { roleId: id }, body: data }),
};
