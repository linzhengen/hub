import { permissionServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-permission-v1-service';
import type { RequestParameters } from '@/api/helper';

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

export const permissionService = {
  listPermissions: (params?: ListPermissionsParams) => apiRequest<ListPermissionResponse>(ops.listPermission, { query: params }),
  getPermission: (id: string) => apiRequest<GetPermissionResponse>(ops.getPermission, { path: { id } }),
  createPermission: (data: CreatePermissionRequest) => apiRequest<CreatePermissionResponse>(ops.createPermission, { body: data }),
  updatePermission: (id: string, data: UpdatePermissionRequest) => apiRequest<UpdatePermissionResponse>(ops.updatePermission, { path: { id }, body: data }),
  deletePermission: (id: string) => apiRequest<DeletePermissionResponse>(ops.deletePermission, { path: { id } }),
};
