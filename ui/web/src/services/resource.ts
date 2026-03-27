import { fetchApi } from '@/lib/api-client';

export type ResourceType = 'TYPE_UNSPECIFIED' | 'TYPE_MENU' | 'TYPE_API';
export type ResourceStatus = 'STATUS_UNSPECIFIED' | 'STATUS_ACTIVE' | 'STATUS_INACTIVE';

export interface Identifier {
  api?: string;
  category?: string;
}

export interface Resource {
  id: string;
  name: string;
  type: ResourceType;
  description?: string;
  status: ResourceStatus;
  identifier?: Identifier;
  metadata?: Record<string, string>;
  component?: string;
  path?: string;
  displayOrder?: number;
  parentId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateResourceRequest {
  name: string;
  type: ResourceType;
  description?: string;
  status?: ResourceStatus;
  identifier?: Identifier;
  metadata?: Record<string, string>;
  component?: string;
  path?: string;
  displayOrder?: number;
  parentId?: string;
}

export interface UpdateResourceRequest {
  name?: string;
  type?: ResourceType;
  description?: string;
  status?: ResourceStatus;
  identifier?: Identifier;
  metadata?: Record<string, string>;
  component?: string;
  path?: string;
  displayOrder?: number;
  parentId?: string;
}

export interface CreateMenuResourceRequest {
  name: string;
  description?: string;
  status?: ResourceStatus;
  metadata?: Record<string, string>;
  component?: string;
  path?: string;
  displayOrder?: number;
  parentId?: string;
}

export interface UpdateMenuResourceRequest {
  name?: string;
  description?: string;
  status?: ResourceStatus;
  metadata?: Record<string, string>;
  component?: string;
  path?: string;
  displayOrder?: number;
  parentId?: string;
}

export interface ListResourceResponse {
  resources: Resource[];
  total: string;
}

export interface GetResourceResponse {
  resource: Resource;
}

export interface CreateResourceResponse {
  resource: Resource;
}

export interface UpdateResourceResponse {
  resource: Resource;
}

export interface DeleteResourceResponse {
  // empty
}

export interface ListMenuResourceResponse {
  resources: Resource[];
  total: string;
}

export interface CreateMenuResourceResponse {
  resource: Resource;
}

export interface UpdateMenuResourceResponse {
  resource: Resource;
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

export const resourceService = {
  listResources: (params?: { limit?: number; offset?: number; resourceIds?: string[]; resourceName?: string; status?: ResourceStatus; resourceType?: ResourceType }) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListResourceResponse>(`/resources${query}`);
  },
  getResource: (id: string) => fetchApi<GetResourceResponse>(`/resources/${id}`),
  createResource: (data: CreateResourceRequest) => fetchApi<CreateResourceResponse>('/resources', { method: 'POST', body: JSON.stringify(data) }),
  updateResource: (id: string, data: UpdateResourceRequest) => fetchApi<UpdateResourceResponse>(`/resources/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteResource: (id: string) => fetchApi<DeleteResourceResponse>(`/resources/${id}`, { method: 'DELETE' }),

  listMenuResources: (params?: { limit?: number; offset?: number; resourceIds?: string[]; resourceName?: string; status?: ResourceStatus }) => {
    const query = params ? buildQueryString(params) : '';
    return fetchApi<ListMenuResourceResponse>(`/resources/menus${query}`);
  },
  createMenuResource: (data: CreateMenuResourceRequest) => fetchApi<CreateMenuResourceResponse>('/resources/menus', { method: 'POST', body: JSON.stringify(data) }),
  updateMenuResource: (id: string, data: UpdateMenuResourceRequest) => fetchApi<UpdateMenuResourceResponse>(`/resources/${id}/menus`, { method: 'PUT', body: JSON.stringify(data) }),
};
