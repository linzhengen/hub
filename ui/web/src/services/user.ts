import { userServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
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
export type SendMeVerifyEmailResponse = components['schemas']['v1SendMeVerifyEmailResponse'];

// Helper type for list users parameters (query)
export type ListUsersParams = RequestParameters<paths, '/api/v1/users', 'get'>;

export const userService = {
  getMe: () => apiRequest<GetMeResponse>(ops.getMe),
  getMeMenus: () => apiRequest<GetMeMenusResponse>(ops.getMeMenus),
  updateMe: (data: UpdateMeRequest) => apiRequest<UpdateMeResponse>(ops.updateMe, { body: data }),
  sendMeVerifyEmail: () => apiRequest<SendMeVerifyEmailResponse>(ops.sendMeVerifyEmail),

  listUsers: (params?: ListUsersParams) => apiRequest<ListUsersResponse>(ops.listUser, { query: params }),
  getUser: (id: string) => apiRequest<GetUserResponse>(ops.getUser, { path: { id } }),
  createUser: (data: CreateUserRequest) => apiRequest<CreateUserResponse>(ops.createUser, { body: data }),
  updateUser: (id: string, data: UpdateUserRequest) => apiRequest<UpdateUserResponse>(ops.updateUser, { path: { id }, body: data }),
  deleteUser: (id: string) => apiRequest<DeleteUserResponse>(ops.deleteUser, { path: { id } }),

  assignGroup: (id: string, groupId: string) => apiRequest<AssignGroupResponse>(ops.assignGroup, { path: { id }, body: { groupId } }),
  unassignGroup: (id: string, groupId: string) => apiRequest<UnassignGroupResponse>(ops.unassignGroup, { path: { id }, body: { groupId } }),
};
