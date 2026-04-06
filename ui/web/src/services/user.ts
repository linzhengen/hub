import { fetchApi } from '@/lib/api-client';
import type { Group } from './group';
import type { ResourceType } from './resource';
import type { paths, components } from '@/api/schema/user-v1-service';
import type { RequestParameters } from '@/api/helper';

// Re-export schema types for backward compatibility
export type UserStatus = components['schemas']['v1UserStatus'];
export type User = components['schemas']['v1User'];
export type Menu = components['schemas']['v1Menu'];
export type MenuMeta = components['schemas']['v1MenuMeta'];
export type ListUsersResponse = components['schemas']['v1ListUserResponse'];
export type GetUserResponse = components['schemas']['v1GetUserResponse'];
export type CreateUserRequest = components['schemas']['v1CreateUserRequest'];
export type CreateUserResponse = components['schemas']['v1CreateUserResponse'];
export type UpdateUserRequest = components['schemas']['UserServiceUpdateUserBody'];
export type UpdateUserResponse = components['schemas']['v1UpdateUserResponse'];
export type DeleteUserResponse = components['schemas']['v1DeleteUserResponse'];
export type AssignGroupResponse = components['schemas']['v1AssignGroupResponse'];
export type UnassignGroupResponse = components['schemas']['v1UnassignGroupResponse'];
export type GetMeMenusResponse = components['schemas']['v1GetMeMenusResponse'];
export type UpdateMeRequest = components['schemas']['v1UpdateMeRequest'];
export type UpdateMeResponse = components['schemas']['v1UpdateMeResponse'];
export type GetMeResponse = components['schemas']['v1GetMeResponse'];

// Helper type for list users parameters (query)
export type ListUsersParams = RequestParameters<paths, '/api/v1/users', 'get'>;

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

export const userService = {
  getMe: () => fetchApi<GetMeResponse>('/me'),
  getMeMenus: () => fetchApi<GetMeMenusResponse>('/me/menus'),
  updateMe: (data: UpdateMeRequest) => fetchApi<UpdateMeResponse>('/me', { method: 'PUT', body: JSON.stringify(data) }),

  listUsers: (params?: ListUsersParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListUsersResponse>(`/users${query}`);
  },
  getUser: (id: string) => fetchApi<GetUserResponse>(`/users/${id}`),
  createUser: (data: CreateUserRequest) => fetchApi<CreateUserResponse>('/users', { method: 'POST', body: JSON.stringify(data) }),
  updateUser: (id: string, data: UpdateUserRequest) => fetchApi<UpdateUserResponse>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteUser: (id: string) => fetchApi<DeleteUserResponse>(`/users/${id}`, { method: 'DELETE' }),

  assignGroup: (id: string, groupId: string) => fetchApi<AssignGroupResponse>(`/users/${id}/groups/assign`, { method: 'POST', body: JSON.stringify({ groupId }) }),
  unassignGroup: (id: string, groupId: string) => fetchApi<UnassignGroupResponse>(`/users/${id}/groups/unassign`, { method: 'POST', body: JSON.stringify({ groupId }) }),
};
