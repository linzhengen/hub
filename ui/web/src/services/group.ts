import { fetchApi } from '@/lib/api-client';
import type { paths, components } from '@/api/schema/system-group-v1-service';
import type {
  UrlPaths,
  RequestParameters,
  RequestData,
  ResponseData,
  HttpMethodsFilteredByPath,
} from '@/api/helper';

// Re-export schema types for backward compatibility
export type GroupStatus = components['schemas']['v1GroupStatus'];
export type Group = components['schemas']['v1Group'];
export type ListGroupsResponse = components['schemas']['v1ListGroupResponse'];
export type GetGroupResponse = components['schemas']['v1GetGroupResponse'];
export type CreateGroupRequest = components['schemas']['v1CreateGroupRequest'];
export type CreateGroupResponse = components['schemas']['v1CreateGroupResponse'];
export type UpdateGroupRequest = components['schemas']['GroupServiceUpdateGroupBody'];
export type UpdateGroupResponse = components['schemas']['v1UpdateGroupResponse'];
export type AssignRoleRequest = components['schemas']['GroupServiceAssignRoleBody'];
export type AssignRoleResponse = components['schemas']['v1AssignRoleResponse'];
export type AssignRolesToGroupRequest = components['schemas']['GroupServiceAssignRolesToGroupBody'];
export type AssignRolesToGroupResponse = components['schemas']['v1AssignRolesToGroupResponse'];
export type AddUsersToGroupRequest = components['schemas']['GroupServiceAddUsersToGroupBody'];
export type AddUsersToGroupResponse = components['schemas']['v1AddUsersToGroupResponse'];
export type RemoveUsersFromGroupRequest = components['schemas']['GroupServiceRemoveUsersFromGroupBody'];
export type RemoveUsersFromGroupResponse = components['schemas']['v1RemoveUsersFromGroupResponse'];
export type DeleteGroupResponse = components['schemas']['v1DeleteGroupResponse'];

// Helper type for list groups parameters (query)
export type ListGroupsParams = RequestParameters<paths, '/api/v1/groups', 'get'>;

// Convert full path to endpoint (remove /api/v1 prefix)
const toEndpoint = (fullPath: string) => fullPath.replace('/api/v1', '');

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
