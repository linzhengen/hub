import { fetchApi } from '@/lib/api-client';

export interface Role {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export const roleService = {
  listRoles: () => fetchApi<{ roles: Role[], total: number }>('/roles'),
  getRole: (id: string) => fetchApi<Role>(`/roles/${id}`),
  createRole: (data: Partial<Role>) => fetchApi<Role>('/roles', { method: 'POST', body: JSON.stringify(data) }),
  updateRole: (id: string, data: Partial<Role>) => fetchApi<Role>(`/roles/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteRole: (id: string) => fetchApi<void>(`/roles/${id}`, { method: 'DELETE' }),

  assignPermission: (id: string, permissionId: string) => fetchApi<void>(`/roles/${id}/permissions/assign`, { method: 'POST', body: JSON.stringify({ permissionId }) }),
  addPermissionsToRole: (id: string, permissionIds: string[]) => fetchApi<void>(`/roles/${id}/permissions`, { method: 'POST', body: JSON.stringify({ permissionIds }) }),
  removePermissionsFromRole: (id: string, permissionIds: string[]) => fetchApi<void>(`/roles/${id}/permissions/remove`, { method: 'POST', body: JSON.stringify({ permissionIds }) }),
};
