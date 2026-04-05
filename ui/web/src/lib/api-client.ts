import { clearTokens } from '@/lib/auth-token';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

/**
 * ローカルストレージからトークンを取得
 */
function getTokenFromStorage(): string | null {
  return localStorage.getItem('keycloak_token');
}

export async function fetchApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  // ローカルストレージからトークンを取得
  const token = getTokenFromStorage();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
    console.log('API Request with token (length):', token.length);
  } else {
    console.warn('API Request without token');
  }

  console.log('API Request:', `${API_BASE_URL}${endpoint}`);

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let message = 'An error occurred';
    let errorData: any = null;

    try {
      errorData = await response.json();
      message = errorData.message || message;
    } catch (e) {
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

      // トークンをクリア（次回のcheck()でKeycloakログイン画面へ）
      clearTokens();

      // 401エラーは特別なエラーとしてスロー
      const error = new Error(message) as any;
      error.status = 401;
      error.data = errorData;
      throw error;
    }

    // その他のエラー
    throw new Error(message);
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}
