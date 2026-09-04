import { getActiveOrgId } from '@/lib/active-org';
import { clearTokens, getToken } from '@/lib/auth-token';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

interface ApiErrorBody {
  message?: string;
  [key: string]: unknown;
}

export class ApiError extends Error {
  status: number;
  data: ApiErrorBody | null;

  constructor(message: string, status: number, data: ApiErrorBody | null) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
  }
}

/**
 * Sends a request and returns the raw {@link Response}.
 *
 * Everything a caller must not get wrong lives here - the bearer token, the
 * error shape, and the 401 handling that clears the session and notifies the
 * UI. Streaming callers need the body rather than a parsed object, so they take
 * this and read `response.body` themselves instead of re-implementing any of it.
 */
export async function fetchApiResponse(endpoint: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  // Which organization the request is about. Sent here rather than added at
  // each call site for the same reason the bearer token is: one place that can
  // be got wrong instead of one per service.
  //
  // Absent when the user has not chosen one, which the server reads as "any
  // organization I hold access in".
  const activeOrgId = getActiveOrgId();
  if (activeOrgId) {
    headers['hub-org'] = activeOrgId;
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
    if (import.meta.env.DEV) {
      console.log('API Request with token (length):', token.length);
    }
  } else if (import.meta.env.DEV) {
    console.warn('API Request without token');
  }

  if (import.meta.env.DEV) {
    console.log('API Request:', `${API_BASE_URL}${endpoint}`);
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let message = 'An error occurred';
    let errorData: ApiErrorBody | null = null;

    try {
      errorData = await response.json();
      message = errorData?.message || message;
    } catch {
      message = response.statusText;
    }

    console.error('API Error:', {
      endpoint,
      status: response.status,
      statusText: response.statusText,
      message,
      hasToken: !!token,
    });

    // 401 Unauthorizedエラーの場合
    if (response.status === 401) {
      console.error('API returned 401 Unauthorized - Clearing tokens and throwing error');

      // トークンをクリア
      clearTokens();

      // カスタムイベントを発火してUIに通知（セッション切れモーダル表示用）
      window.dispatchEvent(new CustomEvent('api-unauthorized'));

      // 401エラーは特別なエラーとしてスロー
      throw new ApiError(message, 401, errorData);
    }

    // その他のエラー
    throw new Error(message);
  }

  return response;
}

export async function fetchApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetchApiResponse(endpoint, options);

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}
