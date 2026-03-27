import { fetchApi } from '@/lib/api-client';
import { Group } from './group';
import { ResourceType } from './resource';

export type UserStatus = 'STATUS_UNSPECIFIED' | 'STATUS_ACTIVE' | 'STATUS_INACTIVE';

export interface User {
  id: string;
  username: string;
  email: string;
  status: UserStatus;
  groupIds: string[];
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  groupIds: string[];
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  password?: string;
  status?: UserStatus;
  groupIds?: string[];
}

export interface UpdateMeRequest {
  username?: string;
  email?: string;
  password?: string;
}

export interface MenuMeta {
  authority?: string;
  badge?: string;
  hideInMenu?: boolean;
  icon?: string;
  keepAlive?: boolean;
  order?: string;
  title?: string;
}

export interface Menu {
  authCode?: string;
  children?: Menu[];
  component?: string;
  identifier?: string;
  meta?: MenuMeta;
  name?: string;
  path?: string;
  type?: ResourceType;
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

export interface ListUsersParams {
  limit?: number;
  offset?: number;
  userIds?: string[];
  userEmails?: string[];
  userName?: string;
  status?: UserStatus;
  groupIds?: string[];
}

export interface GetUserResponse {
  user: User;
}

export interface CreateUserResponse {
  user: User;
}

export interface UpdateUserResponse {
  user: User;
}

export interface DeleteUserResponse {
  // empty
}

export interface AssignGroupResponse {
  user: User;
}

export interface UnassignGroupResponse {
  user: User;
}

export interface ListUsersResponse {
  users: User[];
  total: string;
}

export interface GetMeMenusResponse {
  menus: Menu[];
}

export interface UpdateMeResponse {
  user: User;
  groups: Group[];
}

export interface GetMeResponse {
  user: User;
  groups: Group[];
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
