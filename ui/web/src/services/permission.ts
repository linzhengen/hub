import { fetchApi } from '@/lib/api-client';
import type { paths, components } from '@/api/schema/system-permission-v1-service';
import type {
  RequestParameters,
} from '@/api/helper';

// Re-export schema types for backward compatibility
export type Permission = components['schemas']['v1Permission'];
export type ListPermissionResponse = components['schemas']['v1ListPermissionResponse'];
export type GetPermissionResponse = components['schemas']['v1GetPermissionResponse'];
export type CreatePermissionRequest = components['schemas']['v1CreatePermissionRequest'];
export type CreatePermissionResponse = components['schemas']['v1CreatePermissionResponse'];
export type UpdatePermissionRequest = components['schemas']['PermissionServiceUpdatePermissionBody'];
export type UpdatePermissionResponse = components['schemas']['v1UpdatePermissionResponse'];
export type DeletePermissionResponse = components['schemas']['v1DeletePermissionResponse'];

// Helper type for list permissions parameters (query)
export type ListPermissionsParams = RequestParameters<paths, '/api/v1/permissions', 'get'>;

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
  listPermissions: (params?: ListPermissionsParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListPermissionResponse>(`/permissions${query}`);
  },
  getPermission: (id: string) => fetchApi<GetPermissionResponse>(`/permissions/${id}`),
  createPermission: (data: CreatePermissionRequest) => fetchApi<CreatePermissionResponse>('/permissions', { method: 'POST', body: JSON.stringify(data) }),
  updatePermission: (id: string, data: UpdatePermissionRequest) => fetchApi<UpdatePermissionResponse>(`/permissions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePermission: (id: string) => fetchApi<DeletePermissionResponse>(`/permissions/${id}`, { method: 'DELETE' }),
};
