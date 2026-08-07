import type { ApiOperation } from '@/api/operations';
import { fetchApi } from '@/lib/api-client';

export type QueryScalar = string | number | boolean;
export type QueryValue = QueryScalar | readonly QueryScalar[] | null | undefined;
export type QueryParams = Readonly<Record<string, QueryValue>>;
export type PathParams = Readonly<Record<string, QueryScalar>>;

/**
 * Serialises query parameters the way grpc-gateway parses them.
 *
 * Repeated fields are repeated keys - `?userIds=a&userIds=b` - not one
 * comma-joined value, which the gateway would read as the single id "a,b".
 */
export function buildQueryString(params: QueryParams | undefined): string {
  if (!params) return '';

  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null) continue;
    if (Array.isArray(value)) {
      for (const item of value) searchParams.append(key, String(item));
    } else {
      searchParams.append(key, String(value as QueryScalar));
    }
  }

  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : '';
}

/** Substitutes an operation's `{placeholder}` segments, escaping each value. */
export function buildPath(operation: ApiOperation, params?: PathParams): string {
  return operation.pathParams.reduce((path, name) => {
    const value = params?.[name];
    if (value === undefined || value === null || value === '') {
      throw new Error(`missing path parameter "${name}" for ${operation.method} ${operation.path}`);
    }
    return path.replace(`{${name}}`, encodeURIComponent(String(value)));
  }, operation.path);
}

export interface ApiRequestInit {
  path?: PathParams;
  query?: QueryParams;
  body?: unknown;
}

/**
 * Calls a generated operation. The verb, path and path parameters come from the
 * protobuf definition rather than from a string literal at the call site, so a
 * route that moves in a .proto fails the TypeScript build here instead of at
 * runtime.
 */
export function apiRequest<T>(operation: ApiOperation, init: ApiRequestInit = {}): Promise<T> {
  const url = `${buildPath(operation, init.path)}${buildQueryString(init.query)}`;
  const options: RequestInit = { method: operation.method };

  if (init.body !== undefined) {
    options.body = JSON.stringify(init.body);
  } else if (operation.method === 'POST' || operation.method === 'PUT' || operation.method === 'PATCH') {
    // The gateway rejects a body-mapped request with no body at all.
    options.body = '{}';
  }

  return fetchApi<T>(url, options);
}
