/**
 * 認証トークンのインメモリ管理
 *
 * XSS発生時のトークン窃取リスクを避けるため、localStorageには保存せず
 * モジュールスコープの変数にのみ保持する。ページリロードで失われるが、
 * AuthProvider は Keycloak の SSO セッション（Cookie）を用いて
 * initializeKeycloak() でサイレントに再認証する。
 */

export interface JwtPayload {
  sub?: string;
  exp?: number;
  iat?: number;
  email?: string;
  email_verified?: boolean;
  name?: string;
  given_name?: string;
  family_name?: string;
  preferred_username?: string;
  realm_access?: { roles?: string[] };
  [claim: string]: unknown;
}

let token: string | null = null;
let refreshToken: string | null = null;
let tokenExpiry: number | null = null;

/**
 * トークンをメモリに保存
 */
export function saveToken(newToken: string, newRefreshToken?: string, expiresIn?: number): void {
  if (!newToken) {
    console.warn('Attempted to save empty token');
    return;
  }

  token = newToken;

  if (newRefreshToken) {
    refreshToken = newRefreshToken;
  }

  if (expiresIn) {
    // トークンの有効期限を計算（現在時刻 + expiresIn秒 - 30秒のマージン）
    tokenExpiry = Date.now() + (expiresIn * 1000) - 30000;
  }
}

/**
 * メモリからトークンを取得
 */
export function getToken(): string | null {
  return token;
}

/**
 * メモリからリフレッシュトークンを取得
 */
export function getRefreshToken(): string | null {
  return refreshToken;
}

/**
 * トークンが有効かチェック（期限切れでないか）
 */
export function isTokenValid(): boolean {
  if (!token) {
    return false;
  }

  if (tokenExpiry === null) {
    // 有効期限情報がない場合はとりあえず有効とみなす
    return true;
  }

  return Date.now() < tokenExpiry;
}

/**
 * トークンの有効期限を取得（ミリ秒）
 */
export function getTokenExpiry(): number | null {
  return tokenExpiry;
}

/**
 * トークンが期限切れまであと何秒か取得
 */
export function getSecondsUntilExpiry(): number | null {
  const expiry = getTokenExpiry();
  if (!expiry) {
    return null;
  }

  const now = Date.now();
  const seconds = Math.max(0, Math.floor((expiry - now) / 1000));
  return seconds;
}

/**
 * すべてのトークンをメモリから削除
 */
export function clearTokens(): void {
  token = null;
  refreshToken = null;
  tokenExpiry = null;
}

/**
 * JWTトークンをパースしてペイロードを取得
 */
export function parseToken(token: string): JwtPayload | null {
  try {
    if (!token) return null;

    // JWTトークンは base64url エンコードされた3つの部分に分かれている
    const parts = token.split('.');
    if (parts.length !== 3) {
      console.warn('Invalid JWT token format');
      return null;
    }

    // ペイロード部分をデコード
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded);
  } catch (error) {
    console.error('Failed to parse token:', error);
    return null;
  }
}

/**
 * メモリからトークンを取得してパース
 */
export function getParsedToken(): JwtPayload | null {
  const currentToken = getToken();
  if (!currentToken) return null;
  return parseToken(currentToken);
}

/**
 * トークン情報をデバッグ用に表示
 */
export function debugTokenInfo(): void {
  const token = getToken();
  const refreshToken = getRefreshToken();
  const expiry = getTokenExpiry();
  const isValid = isTokenValid();
  const secondsUntilExpiry = getSecondsUntilExpiry();
  const parsedToken = getParsedToken();

  console.log('Token Debug Info:');
  console.log('- Token exists:', !!token);
  console.log('- Token length:', token?.length || 0);
  console.log('- Refresh token exists:', !!refreshToken);
  console.log('- Token valid:', isValid);
  console.log('- Expiry time:', expiry ? new Date(expiry).toISOString() : 'N/A');
  console.log('- Seconds until expiry:', secondsUntilExpiry);
  console.log('- Parsed token:', parsedToken);
}
