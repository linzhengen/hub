import { fetchApi } from '@/lib/api-client';
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

export const resourceService = {
  listResources: (params?: ListResourcesParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListResourceResponse>(`/resources${query}`);
  },
  getResource: (id: string) => fetchApi<GetResourceResponse>(`/resources/${id}`),
  createResource: (data: CreateResourceRequest) => fetchApi<CreateResourceResponse>('/resources', { method: 'POST', body: JSON.stringify(data) }),
  updateResource: (id: string, data: UpdateResourceRequest) => fetchApi<UpdateResourceResponse>(`/resources/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteResource: (id: string) => fetchApi<DeleteResourceResponse>(`/resources/${id}`, { method: 'DELETE' }),

  listMenuResources: (params?: ListMenuResourcesParams) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListMenuResourceResponse>(`/resources/menus${query}`);
  },
  createMenuResource: (data: CreateMenuResourceRequest) => fetchApi<CreateMenuResourceResponse>('/resources/menus', { method: 'POST', body: JSON.stringify(data) }),
  updateMenuResource: (id: string, data: UpdateMenuResourceRequest) => fetchApi<UpdateMenuResourceResponse>(`/resources/${id}/menus`, { method: 'PUT', body: JSON.stringify(data) }),
};
