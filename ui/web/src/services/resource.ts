import { fetchApi } from '@/lib/api-client';

export interface Resource {
  id: string;
  name: string;
  type: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MenuResource {
  id: string;
  name: string;
  path: string;
  icon?: string;
  parentId?: string;
  order: number;
}

export const resourceService = {
  listResources: () => fetchApi<{ resources: Resource[], total: number }>('/resources'),
  getResource: (id: string) => fetchApi<Resource>(`/resources/${id}`),
  createResource: (data: Partial<Resource>) => fetchApi<Resource>('/resources', { method: 'POST', body: JSON.stringify(data) }),
  updateResource: (id: string, data: Partial<Resource>) => fetchApi<Resource>(`/resources/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteResource: (id: string) => fetchApi<void>(`/resources/${id}`, { method: 'DELETE' }),

  listMenuResources: () => fetchApi<{ menus: MenuResource[], total: number }>('/resources/menus'),
  createMenuResource: (data: Partial<MenuResource>) => fetchApi<MenuResource>('/resources/menus', { method: 'POST', body: JSON.stringify(data) }),
  updateMenuResource: (id: string, data: Partial<MenuResource>) => fetchApi<MenuResource>(`/resources/${id}/menus`, { method: 'PUT', body: JSON.stringify(data) }),
};
