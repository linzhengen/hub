import type { ApiOperation } from '@/api/operations';
import { fetchApi, fetchApiResponse } from '@/lib/api-client';

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
  /** Aborts the request - pass a controller's signal to cancel it. */
  signal?: AbortSignal;
}

/**
 * Calls a generated operation. The verb, path and path parameters come from the
 * protobuf definition rather than from a string literal at the call site, so a
 * route that moves in a .proto fails the TypeScript build here instead of at
 * runtime.
 */
export function apiRequest<T>(operation: ApiOperation, init: ApiRequestInit = {}): Promise<T> {
  return fetchApi<T>(buildUrl(operation, init), buildOptions(operation, init));
}

function buildUrl(operation: ApiOperation, init: ApiRequestInit): string {
  return `${buildPath(operation, init.path)}${buildQueryString(init.query)}`;
}

function buildOptions(operation: ApiOperation, init: ApiRequestInit): RequestInit {
  const options: RequestInit = { method: operation.method };

  if (init.signal) options.signal = init.signal;

  if (init.body !== undefined) {
    options.body = JSON.stringify(init.body);
  } else if (operation.method === 'POST' || operation.method === 'PUT' || operation.method === 'PATCH') {
    // The gateway rejects a body-mapped request with no body at all.
    options.body = '{}';
  }

  return options;
}

/** An error the server sent partway through a stream, after the 200 headers. */
export class ApiStreamError extends Error {
  readonly code?: number;
  readonly httpStatus?: number;

  constructor(message: string, code?: number, httpStatus?: number) {
    super(message);
    this.name = 'ApiStreamError';
    this.code = code;
    this.httpStatus = httpStatus;
  }
}

/** The envelope grpc-gateway wraps each message of a server-streaming rpc in. */
interface StreamFrame<T> {
  result?: T;
  error?: { message?: string; grpc_code?: number; http_code?: number };
}

/**
 * Calls a server-streaming operation, yielding each message as it arrives.
 *
 * grpc-gateway sends a streaming response as newline-delimited JSON, one
 * `{"result": ...}` object per message, so waiting for the whole body - which is
 * what {@link apiRequest} does - would hold every message back until the last
 * one. Reading it incrementally is what lets the chat UI fill in as the model
 * answers rather than after it has finished.
 *
 * A failure partway through arrives as a `{"error": ...}` frame on an otherwise
 * successful response; that is thrown as an {@link ApiStreamError}, so a caller
 * handles it the same way as a request that never started.
 *
 * Leaving the loop early - `break`, or an exception in the body - cancels the
 * response, so an unfinished answer does not keep streaming into nothing. Pass
 * `init.signal` to cancel from outside.
 */
export async function* apiStream<T>(
  operation: ApiOperation,
  init: ApiRequestInit = {},
): AsyncGenerator<T, void, undefined> {
  const response = await fetchApiResponse(buildUrl(operation, init), buildOptions(operation, init));
  if (!response.body) {
    throw new Error(`${operation.method} ${operation.path} returned no body to stream`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      // A chunk can carry several messages, or half of one. Only whole lines
      // are parsed; the remainder stays buffered for the next chunk.
      let newline = buffer.indexOf('\n');
      while (newline >= 0) {
        const line = buffer.slice(0, newline);
        buffer = buffer.slice(newline + 1);
        const message = parseFrame<T>(line);
        if (message !== undefined) yield message;
        newline = buffer.indexOf('\n');
      }
    }

    // A last message with no trailing newline.
    const message = parseFrame<T>(buffer + decoder.decode());
    if (message !== undefined) yield message;
  } finally {
    await reader.cancel().catch(() => undefined);
  }
}

function parseFrame<T>(line: string): T | undefined {
  const trimmed = line.trim();
  if (trimmed === '') return undefined;

  let frame: StreamFrame<T>;
  try {
    frame = JSON.parse(trimmed) as StreamFrame<T>;
  } catch {
    throw new ApiStreamError(`unparseable frame in stream: ${trimmed}`);
  }

  if (frame.error) {
    throw new ApiStreamError(
      frame.error.message ?? 'the stream failed',
      frame.error.grpc_code,
      frame.error.http_code,
    );
  }
  return frame.result;
}
