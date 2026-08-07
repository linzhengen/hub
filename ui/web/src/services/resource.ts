import { resourceServiceOperations as ops } from '@/api/operations';
import { apiRequest } from '@/api/request';
import type { paths, components } from '@/api/schema/system-resource-v1-service';
import type { RequestParameters } from '@/api/helper';

// Re-export schema types for backward compatibility
export type ResourceType = components['schemas']['resourceV1Type'];
export type ResourceStatus = components['schemas']['resourceV1Status'];
export type Identifier = components['schemas']['v1Identifier'];
export type Resource = components['schemas']['v1Resource'];
export type ListResourceResponse = components['schemas']['v1ListResourceResponse'];
export type GetResourceResponse = components['schemas']['v1GetResourceResponse'];
export type CreateResourceRequest = components['schemas']['v1CreateResourceRequest'];
export type CreateResourceResponse = components['schemas']['v1CreateResourceResponse'];
export type UpdateResourceRequest = components['schemas']['ResourceServiceUpdateResourceBody'];
export type UpdateResourceResponse = components['schemas']['v1UpdateResourceResponse'];
export type DeleteResourceResponse = components['schemas']['v1DeleteResourceResponse'];
export type ListMenuResourceResponse = components['schemas']['v1ListMenuResourceResponse'];
export type CreateMenuResourceRequest = components['schemas']['v1CreateMenuResourceRequest'];
export type CreateMenuResourceResponse = components['schemas']['v1CreateMenuResourceResponse'];
export type UpdateMenuResourceRequest = components['schemas']['ResourceServiceUpdateMenuResourceBody'];
export type UpdateMenuResourceResponse = components['schemas']['v1UpdateMenuResourceResponse'];

// Helper types for parameters
export type ListResourcesParams = RequestParameters<paths, '/api/v1/resources', 'get'>;
export type ListMenuResourcesParams = RequestParameters<paths, '/api/v1/resources/menus', 'get'>;

export const resourceService = {
  listResources: (params?: ListResourcesParams) => apiRequest<ListResourceResponse>(ops.listResource, { query: params }),
  getResource: (id: string) => apiRequest<GetResourceResponse>(ops.getResource, { path: { id } }),
  createResource: (data: CreateResourceRequest) => apiRequest<CreateResourceResponse>(ops.createResource, { body: data }),
  updateResource: (id: string, data: UpdateResourceRequest) => apiRequest<UpdateResourceResponse>(ops.updateResource, { path: { id }, body: data }),
  deleteResource: (id: string) => apiRequest<DeleteResourceResponse>(ops.deleteResource, { path: { id } }),

  listMenuResources: (params?: ListMenuResourcesParams) => apiRequest<ListMenuResourceResponse>(ops.listMenuResource, { query: params }),
  createMenuResource: (data: CreateMenuResourceRequest) => apiRequest<CreateMenuResourceResponse>(ops.createMenuResource, { body: data }),
  updateMenuResource: (id: string, data: UpdateMenuResourceRequest) => apiRequest<UpdateMenuResourceResponse>(ops.updateMenuResource, { path: { id }, body: data }),
};
