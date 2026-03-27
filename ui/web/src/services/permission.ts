import { fetchApi } from '@/lib/api-client';

export interface Permission {
  id: string;
  name: string;
  resourceId: string;
  verb: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePermissionRequest {
  resourceId: string;
  verb: string;
  description?: string;
}

export interface UpdatePermissionRequest {
  verb?: string;
  description?: string;
}

export interface ListPermissionResponse {
  permissions: Permission[];
  total: string;
}

export interface GetPermissionResponse {
  permission: Permission;
}

export interface CreatePermissionResponse {
  permission: Permission;
}

export interface UpdatePermissionResponse {
  permission: Permission;
}

export interface DeletePermissionResponse {
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

export const permissionService = {
  listPermissions: (params?: { limit?: number; offset?: number; permissionIds?: string[]; permissionName?: string }) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListPermissionResponse>(`/permissions${query}`);
  },
  getPermission: (id: string) => fetchApi<GetPermissionResponse>(`/permissions/${id}`),
  createPermission: (data: CreatePermissionRequest) => fetchApi<CreatePermissionResponse>('/permissions', { method: 'POST', body: JSON.stringify(data) }),
  updatePermission: (id: string, data: UpdatePermissionRequest) => fetchApi<UpdatePermissionResponse>(`/permissions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePermission: (id: string) => fetchApi<DeletePermissionResponse>(`/permissions/${id}`, { method: 'DELETE' }),
};
