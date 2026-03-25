import { fetchApi } from '@/lib/api-client';

export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export const permissionService = {
  listPermissions: () => fetchApi<{ permissions: Permission[], total: number }>('/permissions'),
  getPermission: (id: string) => fetchApi<Permission>(`/permissions/${id}`),
  createPermission: (data: Partial<Permission>) => fetchApi<Permission>('/permissions', { method: 'POST', body: JSON.stringify(data) }),
  updatePermission: (id: string, data: Partial<Permission>) => fetchApi<Permission>(`/permissions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePermission: (id: string) => fetchApi<void>(`/permissions/${id}`, { method: 'DELETE' }),
};
