import { fetchApi } from '@/lib/api-client';
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

function buildQueryString(params: Record<string, any>): string {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null) continue;
    if (Array.isArray(value)) {
      // For arrays, join with commas (common API pattern)
      searchParams.append(key, value.join(','));
    } else {
      searchParams.append(key, value.toString());
    }
  }
  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : '';
}

export const roleService = {
  listRoles: (params?: ListRolesParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListRolesResponse>(`/roles${query}`);
  },
  getRole: (id: string) => fetchApi<GetRoleResponse>(`/roles/${id}`),
  createRole: (data: CreateRoleRequest) => fetchApi<CreateRoleResponse>('/roles', { method: 'POST', body: JSON.stringify(data) }),
  updateRole: (id: string, data: UpdateRoleRequest) => fetchApi<UpdateRoleResponse>(`/roles/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteRole: (id: string) => fetchApi<DeleteRoleResponse>(`/roles/${id}`, { method: 'DELETE' }),

  assignPermission: (id: string, data: AssignPermissionRequest) => fetchApi<AssignPermissionResponse>(`/roles/${id}/permissions/assign`, { method: 'POST', body: JSON.stringify(data) }),
  addPermissionsToRole: (id: string, data: AddPermissionsToRoleRequest) => fetchApi<AddPermissionsToRoleResponse>(`/roles/${id}/permissions`, { method: 'POST', body: JSON.stringify(data) }),
  removePermissionsFromRole: (id: string, data: RemovePermissionsFromRoleRequest) => fetchApi<RemovePermissionsFromRoleResponse>(`/roles/${id}/permissions/remove`, { method: 'POST', body: JSON.stringify(data) }),
};
