import { organizationServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-organization-v1-service';
import type { RequestParameters } from '@/api/helper';

export type Organization = components['schemas']['v1Organization'];
export type OrganizationKind = components['schemas']['v1OrganizationKind'];
export type OrganizationStatus = components['schemas']['v1OrganizationStatus'];
export type CreateOrganizationRequest = components['schemas']['v1CreateOrganizationRequest'];
export type CreateOrganizationResponse = components['schemas']['v1CreateOrganizationResponse'];
export type UpdateOrganizationBody =
  components['schemas']['OrganizationServiceUpdateOrganizationBody'];
export type UpdateOrganizationResponse = components['schemas']['v1UpdateOrganizationResponse'];
export type ListOrganizationResponse = components['schemas']['v1ListOrganizationResponse'];
export type ListMyOrganizationsResponse =
  components['schemas']['v1ListMyOrganizationsResponse'];
export type GetOrganizationResponse = components['schemas']['v1GetOrganizationResponse'];
export type DeleteOrganizationResponse = components['schemas']['v1DeleteOrganizationResponse'];

export type ListOrganizationParams = RequestParameters<paths, '/api/v1/organizations', 'get'>;

/**
 * `listMine` is the switcher's source and needs no permission: it is derived
 * from the caller's own group memberships. `list` is the administrator's view
 * of every tenant and is permission-gated like the rest of the service.
 */
export const organizationService = {
  list: (params?: ListOrganizationParams) =>
    apiRequest<ListOrganizationResponse>(ops.listOrganization, { query: params }),

  listMine: () => apiRequest<ListMyOrganizationsResponse>(ops.listMyOrganizations),

  get: (id: string) => apiRequest<GetOrganizationResponse>(ops.getOrganization, { path: { id } }),

  create: (data: CreateOrganizationRequest) =>
    apiRequest<CreateOrganizationResponse>(ops.createOrganization, { body: data }),

  update: (id: string, data: UpdateOrganizationBody) =>
    apiRequest<UpdateOrganizationResponse>(ops.updateOrganization, { path: { id }, body: data }),

  remove: (id: string) =>
    apiRequest<DeleteOrganizationResponse>(ops.deleteOrganization, { path: { id } }),
};
