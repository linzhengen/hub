import { fetchApi } from '@/lib/api-client';

export type GroupStatus = 'STATUS_UNSPECIFIED' | 'STATUS_ACTIVE' | 'STATUS_INACTIVE';

export interface Group {
  id: string;
  name: string;
  description?: string;
  status: GroupStatus;
  roleIds: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ListGroupsResponse {
  groups: Group[];
  total: string;
}

export interface GetGroupResponse {
  group: Group;
}

export interface CreateGroupRequest {
  name: string;
  description?: string;
  status?: GroupStatus;
}

export interface CreateGroupResponse {
  group: Group;
}

export interface UpdateGroupRequest {
  name?: string;
  description?: string;
  status?: GroupStatus;
}

export interface UpdateGroupResponse {
  group: Group;
}

export interface AssignRoleRequest {
  roleId: string;
}

export interface AssignRoleResponse {
  group: Group;
}

export interface AssignRolesToGroupRequest {
  roleIds: string[];
}

export interface AddUsersToGroupRequest {
  userIds: string[];
}

export interface RemoveUsersFromGroupRequest {
  userIds: string[];
}

export interface ListGroupsParams {
  limit?: number;
  offset?: number;
  groupIds?: string[];
  groupName?: string;
  status?: GroupStatus;
  roleIds?: string[];
}

export interface DeleteGroupResponse {
  // empty
}

export interface AssignRolesToGroupResponse {
  // empty
}

export interface AddUsersToGroupResponse {
  // empty
}

export interface RemoveUsersFromGroupResponse {
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

export const groupService = {
  listGroups: (params?: ListGroupsParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListGroupsResponse>(`/groups${query}`);
  },
  getGroup: (id: string) => fetchApi<GetGroupResponse>(`/groups/${id}`),
  createGroup: (data: CreateGroupRequest) => fetchApi<CreateGroupResponse>('/groups', { method: 'POST', body: JSON.stringify(data) }),
  updateGroup: (id: string, data: UpdateGroupRequest) => fetchApi<UpdateGroupResponse>(`/groups/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteGroup: (id: string) => fetchApi<DeleteGroupResponse>(`/groups/${id}`, { method: 'DELETE' }),

  assignRole: (id: string, data: AssignRoleRequest) => fetchApi<AssignRoleResponse>(`/groups/${id}/roles/assign`, { method: 'POST', body: JSON.stringify(data) }),
  assignRolesToGroup: (id: string, data: AssignRolesToGroupRequest) => fetchApi<AssignRolesToGroupResponse>(`/groups/${id}/roles/assignRoles`, { method: 'POST', body: JSON.stringify(data) }),
  addUsersToGroup: (id: string, data: AddUsersToGroupRequest) => fetchApi<AddUsersToGroupResponse>(`/groups/${id}/users/add`, { method: 'POST', body: JSON.stringify(data) }),
  removeUsersFromGroup: (id: string, data: RemoveUsersFromGroupRequest) => fetchApi<RemoveUsersFromGroupResponse>(`/groups/${id}/users/remove`, { method: 'POST', body: JSON.stringify(data) }),
};
