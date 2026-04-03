import { DataProvider } from '@refinedev/core';
import { fetchApi, API_BASE_URL } from '@/lib/api-client';

// APIレスポンスの型定義
interface ApiResponse<T = any> {
  data: T;
  total?: number;
  page?: number;
  limit?: number;
}

// クエリパラメータをURLに変換するヘルパー関数
function buildQueryString(params: Record<string, any> = {}): string {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      if (Array.isArray(value)) {
        value.forEach(v => searchParams.append(key, v.toString()));
      } else {
        searchParams.append(key, value.toString());
      }
    }
  });

  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : '';
}

// フィルターをAPIパラメータに変換する関数
function convertFiltersToParams(filters: any[] = []): Record<string, any> {
  const params: Record<string, any> = {};

  filters.forEach(filter => {
    if ('field' in filter && 'operator' in filter && 'value' in filter) {
      const { field, operator, value } = filter;

      switch (operator) {
        case 'eq':
          params[`filter[${field}]`] = value;
          break;
        case 'ne':
          params[`filter[${field}][ne]`] = value;
          break;
        case 'lt':
          params[`filter[${field}][lt]`] = value;
          break;
        case 'gt':
          params[`filter[${field}][gt]`] = value;
          break;
        case 'lte':
          params[`filter[${field}][lte]`] = value;
          break;
        case 'gte':
          params[`filter[${field}][gte]`] = value;
          break;
        case 'in':
          params[`filter[${field}][in]`] = Array.isArray(value) ? value.join(',') : value;
          break;
        case 'nin':
          params[`filter[${field}][nin]`] = Array.isArray(value) ? value.join(',') : value;
          break;
        case 'contains':
          params[`filter[${field}][contains]`] = value;
          break;
        default:
          params[`filter[${field}]`] = value;
      }
    }
  });

  return params;
}

// ソートをAPIパラメータに変換する関数
function convertSortersToParams(sorters: any[] = []): Record<string, any> {
  const params: Record<string, any> = {};

  if (sorters.length > 0) {
    const sortFields = sorters.map(sorter => {
      const field = sorter.field;
      const order = sorter.order === 'desc' ? '-' : '';
      return `${order}${field}`;
    });

    params.sort = sortFields.join(',');
  }

  return params;
}

export const dataProvider: DataProvider = {
  getList: async ({ resource, pagination, filters, sorters, meta }) => {
    const { current = 1, pageSize = 10, mode = 'server' } = pagination || {} as any;

    // クエリパラメータの構築
    const params: Record<string, any> = {};

    if (mode === 'server') {
      params.page = current;
      params.limit = pageSize;
    }

    // フィルターの適用
    Object.assign(params, convertFiltersToParams(filters));

    // ソートの適用
    Object.assign(params, convertSortersToParams(sorters));

    // API呼び出し
    const queryString = buildQueryString(params);
    const endpoint = `/${resource}${queryString}`;

    try {
      const response = await fetchApi<ApiResponse>(endpoint);

      return {
        data: response.data || [],
        total: response.total || (Array.isArray(response.data) ? response.data.length : 0),
      };
    } catch (error) {
      console.error(`Error fetching ${resource}:`, error);
      throw error;
    }
  },

  getOne: async ({ resource, id, meta }) => {
    try {
      const response = await fetchApi<ApiResponse>(`/${resource}/${id}`);
      return { data: response.data };
    } catch (error) {
      console.error(`Error fetching ${resource}/${id}:`, error);
      throw error;
    }
  },

  create: async ({ resource, variables, meta }) => {
    try {
      const response = await fetchApi<ApiResponse>(`/${resource}`, {
        method: 'POST',
        body: JSON.stringify(variables),
      });
      return { data: response.data };
    } catch (error) {
      console.error(`Error creating ${resource}:`, error);
      throw error;
    }
  },

  update: async ({ resource, id, variables, meta }) => {
    try {
      const response = await fetchApi<ApiResponse>(`/${resource}/${id}`, {
        method: 'PUT',
        body: JSON.stringify(variables),
      });
      return { data: response.data };
    } catch (error) {
      console.error(`Error updating ${resource}/${id}:`, error);
      throw error;
    }
  },

  deleteOne: async ({ resource, id, meta }) => {
    try {
      await fetchApi(`/${resource}/${id}`, {
        method: 'DELETE',
      });
      return { data: { id } as any };
    } catch (error) {
      console.error(`Error deleting ${resource}/${id}:`, error);
      throw error;
    }
  },

  getApiUrl: () => API_BASE_URL,

  // オプション: カスタムメソッド
  custom: async ({ url, method, filters, sorters, payload, query, headers, meta }) => {
    try {
      // クエリパラメータの構築
      const params: Record<string, any> = {};

      // フィルターの適用
      Object.assign(params, convertFiltersToParams(filters));

      // ソートの適用
      Object.assign(params, convertSortersToParams(sorters));

      // 追加のクエリパラメータ
      if (query) {
        Object.assign(params, query);
      }

      const queryString = buildQueryString(params);
      const fullUrl = `${url}${queryString}`;

      const response = await fetchApi(fullUrl, {
        method: method || 'GET',
        headers: headers as Record<string, string>,
        body: payload ? JSON.stringify(payload) : undefined,
      });

      return { data: response as any };
    } catch (error) {
      console.error('Error in custom request:', error);
      throw error;
    }
  },

  // オプション: 複数取得
  getMany: async ({ resource, ids, meta }) => {
    try {
      const params = { filter: { id: { in: ids.join(',') } } };
      const queryString = buildQueryString(params);
      const response = await fetchApi<ApiResponse>(`/${resource}${queryString}`);

      return { data: response.data || [] };
    } catch (error) {
      console.error(`Error fetching many ${resource}:`, error);
      throw error;
    }
  },

  // オプション: 複数更新
  updateMany: async ({ resource, ids, variables, meta }) => {
    const results = await Promise.all(
      ids.map(id =>
        fetchApi<ApiResponse>(`/${resource}/${id}`, {
          method: 'PUT',
          body: JSON.stringify(variables),
        })
      )
    );

    return { data: results.map(r => r.data) };
  },

  // オプション: 複数削除
  deleteMany: async ({ resource, ids, meta }) => {
    await Promise.all(
      ids.map(id =>
        fetchApi(`/${resource}/${id}`, {
          method: 'DELETE',
        })
      )
    );

    return { data: [] };
  },

  // オプション: 一覧作成
  createMany: async ({ resource, variables, meta }) => {
    const results = await Promise.all(
      variables.map(variable =>
        fetchApi<ApiResponse>(`/${resource}`, {
          method: 'POST',
          body: JSON.stringify(variable),
        })
      )
    );

    return { data: results.map(r => r.data) };
  },
};
