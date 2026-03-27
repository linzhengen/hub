import { fetchApi } from '@/lib/api-client';

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissionIds: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ListRolesResponse {
  roles: Role[];
  total: string;
}

export interface GetRoleResponse {
  role: Role;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
}

export interface CreateRoleResponse {
  role: Role;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

export interface UpdateRoleResponse {
  role: Role;
}

export interface AssignPermissionRequest {
  permissionId: string;
}

export interface AssignPermissionResponse {
  role: Role;
}

export interface AddPermissionsToRoleRequest {
  permissionIds: string[];
}

export interface RemovePermissionsFromRoleRequest {
  permissionIds: string[];
}

export interface ListRolesParams {
  limit?: number;
  offset?: number;
  roleIds?: string[];
  roleName?: string;
  permissionIds?: string[];
}

export interface DeleteRoleResponse {
  // empty
}

export interface AddPermissionsToRoleResponse {
  // empty
}

export interface RemovePermissionsFromRoleResponse {
  // empty
}

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
