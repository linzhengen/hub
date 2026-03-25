import { fetchApi } from '@/lib/api-client';

export interface Group {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export const groupService = {
  listGroups: () => fetchApi<{ groups: Group[], total: number }>('/groups'),
  getGroup: (id: string) => fetchApi<Group>(`/groups/${id}`),
  createGroup: (data: Partial<Group>) => fetchApi<Group>('/groups', { method: 'POST', body: JSON.stringify(data) }),
  updateGroup: (id: string, data: Partial<Group>) => fetchApi<Group>(`/groups/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteGroup: (id: string) => fetchApi<void>(`/groups/${id}`, { method: 'DELETE' }),

  assignRole: (id: string, roleId: string) => fetchApi<void>(`/groups/${id}/roles/assign`, { method: 'POST', body: JSON.stringify({ roleId }) }),
  assignRolesToGroup: (id: string, roleIds: string[]) => fetchApi<void>(`/groups/${id}/roles/assignRoles`, { method: 'POST', body: JSON.stringify({ roleIds }) }),
  addUsersToGroup: (id: string, userIds: string[]) => fetchApi<void>(`/groups/${id}/users/add`, { method: 'POST', body: JSON.stringify({ userIds }) }),
  removeUsersFromGroup: (id: string, userIds: string[]) => fetchApi<void>(`/groups/${id}/users/remove`, { method: 'POST', body: JSON.stringify({ userIds }) }),
};
