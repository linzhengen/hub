import { fetchApi } from '@/lib/api-client';

export interface User {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export const userService = {
  getMe: () => fetchApi<User>('/me'),
  getMeMenus: () => fetchApi<any[]>('/me/menus'),
  updateMe: (data: Partial<User>) => fetchApi<User>('/me', { method: 'PUT', body: JSON.stringify(data) }),

  listUsers: () => fetchApi<{ users: User[], total: number }>('/users'),
  getUser: (id: string) => fetchApi<User>(`/users/${id}`),
  createUser: (data: Partial<User>) => fetchApi<User>('/users', { method: 'POST', body: JSON.stringify(data) }),
  updateUser: (id: string, data: Partial<User>) => fetchApi<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteUser: (id: string) => fetchApi<void>(`/users/${id}`, { method: 'DELETE' }),

  assignGroup: (id: string, groupId: string) => fetchApi<void>(`/users/${id}/groups/assign`, { method: 'POST', body: JSON.stringify({ groupId }) }),
  unassignGroup: (id: string, groupId: string) => fetchApi<void>(`/users/${id}/groups/unassign`, { method: 'POST', body: JSON.stringify({ groupId }) }),
};
