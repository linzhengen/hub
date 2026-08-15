import { groupServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-group-v1-service';
import type { RequestParameters } from '@/api/helper';

// Re-export schema types for backward compatibility
export type GroupStatus = components['schemas']['v1GroupStatus'];
export type Group = components['schemas']['v1Group'];
/** One role the group holds, with when the grant ends. */
export type RoleGrant = components['schemas']['v1RoleGrant'];
export type ListGroupsResponse = components['schemas']['v1ListGroupResponse'];
export type GetGroupResponse = components['schemas']['v1GetGroupResponse'];
export type CreateGroupRequest = components['schemas']['v1CreateGroupRequest'];
export type CreateGroupResponse = components['schemas']['v1CreateGroupResponse'];
export type UpdateGroupRequest = components['schemas']['GroupServiceUpdateGroupBody'];
export type UpdateGroupResponse = components['schemas']['v1UpdateGroupResponse'];
export type AddRolesToGroupRequest = components['schemas']['GroupServiceAddRolesToGroupBody'];
export type AddRolesToGroupResponse = components['schemas']['v1AddRolesToGroupResponse'];
export type RemoveRolesFromGroupRequest = components['schemas']['GroupServiceRemoveRolesFromGroupBody'];
export type RemoveRolesFromGroupResponse = components['schemas']['v1RemoveRolesFromGroupResponse'];
export type AddUsersToGroupRequest = components['schemas']['GroupServiceAddUsersToGroupBody'];
export type AddUsersToGroupResponse = components['schemas']['v1AddUsersToGroupResponse'];
export type RemoveUsersFromGroupRequest = components['schemas']['GroupServiceRemoveUsersFromGroupBody'];
export type RemoveUsersFromGroupResponse = components['schemas']['v1RemoveUsersFromGroupResponse'];
export type DeleteGroupResponse = components['schemas']['v1DeleteGroupResponse'];

// Helper type for list groups parameters (query)
export type ListGroupsParams = RequestParameters<paths, '/api/v1/groups', 'get'>;

export const groupService = {
  listGroups: (params?: ListGroupsParams) => apiRequest<ListGroupsResponse>(ops.listGroup, { query: params }),
  getGroup: (id: string) => apiRequest<GetGroupResponse>(ops.getGroup, { path: { id } }),
  createGroup: (data: CreateGroupRequest) => apiRequest<CreateGroupResponse>(ops.createGroup, { body: data }),
  updateGroup: (id: string, data: UpdateGroupRequest) => apiRequest<UpdateGroupResponse>(ops.updateGroup, { path: { id }, body: data }),
  deleteGroup: (id: string) => apiRequest<DeleteGroupResponse>(ops.deleteGroup, { path: { id } }),

  addRoles: (id: string, data: AddRolesToGroupRequest) => apiRequest<AddRolesToGroupResponse>(ops.addRolesToGroup, { path: { id }, body: data }),
  removeRoles: (id: string, data: RemoveRolesFromGroupRequest) => apiRequest<RemoveRolesFromGroupResponse>(ops.removeRolesFromGroup, { path: { id }, body: data }),
  addUsers: (id: string, data: AddUsersToGroupRequest) => apiRequest<AddUsersToGroupResponse>(ops.addUsersToGroup, { path: { id }, body: data }),
  removeUsers: (id: string, data: RemoveUsersFromGroupRequest) => apiRequest<RemoveUsersFromGroupResponse>(ops.removeUsersFromGroup, { path: { id }, body: data }),
};
