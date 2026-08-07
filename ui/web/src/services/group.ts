import { groupServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-group-v1-service';
import type { RequestParameters } from '@/api/helper';

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

export const groupService = {
  listGroups: (params?: ListGroupsParams) => apiRequest<ListGroupsResponse>(ops.listGroup, { query: params }),
  getGroup: (id: string) => apiRequest<GetGroupResponse>(ops.getGroup, { path: { id } }),
  createGroup: (data: CreateGroupRequest) => apiRequest<CreateGroupResponse>(ops.createGroup, { body: data }),
  updateGroup: (id: string, data: UpdateGroupRequest) => apiRequest<UpdateGroupResponse>(ops.updateGroup, { path: { id }, body: data }),
  deleteGroup: (id: string) => apiRequest<DeleteGroupResponse>(ops.deleteGroup, { path: { id } }),

  assignRole: (id: string, data: AssignRoleRequest) => apiRequest<AssignRoleResponse>(ops.assignRole, { path: { id }, body: data }),
  assignRolesToGroup: (id: string, data: AssignRolesToGroupRequest) => apiRequest<AssignRolesToGroupResponse>(ops.assignRolesToGroup, { path: { groupId: id }, body: data }),
  addUsersToGroup: (id: string, data: AddUsersToGroupRequest) => apiRequest<AddUsersToGroupResponse>(ops.addUsersToGroup, { path: { groupId: id }, body: data }),
  removeUsersFromGroup: (id: string, data: RemoveUsersFromGroupRequest) => apiRequest<RemoveUsersFromGroupResponse>(ops.removeUsersFromGroup, { path: { groupId: id }, body: data }),
};
