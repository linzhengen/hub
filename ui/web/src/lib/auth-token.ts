/**
 * 認証トークンのローカルストレージ管理
 */

const TOKEN_KEY = 'keycloak_token';
const TOKEN_EXPIRY_KEY = 'keycloak_token_expiry';
const REFRESH_TOKEN_KEY = 'keycloak_refresh_token';

/**
 * トークンをローカルストレージに保存
 */
export function saveToken(token: string, refreshToken?: string, expiresIn?: number): void {
  if (!token) {
    console.warn('Attempted to save empty token');
    return;
  }

  localStorage.setItem(TOKEN_KEY, token);

  if (refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }

  if (expiresIn) {
    // トークンの有効期限を計算（現在時刻 + expiresIn秒 - 30秒のマージン）
    const expiryTime = Date.now() + (expiresIn * 1000) - 30000;
    localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
  }
}

/**
 * ローカルストレージからトークンを取得
 */
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * ローカルストレージからリフレッシュトークンを取得
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

/**
 * トークンが有効かチェック（期限切れでないか）
 */
export function isTokenValid(): boolean {
  const token = getToken();
  if (!token) {
    return false;
  }

  const expiryStr = localStorage.getItem(TOKEN_EXPIRY_KEY);
  if (!expiryStr) {
    // 有効期限情報がない場合はとりあえず有効とみなす
    return true;
  }

  const expiryTime = parseInt(expiryStr, 10);
  const now = Date.now();

  return now < expiryTime;
}

/**
 * トークンの有効期限を取得（ミリ秒）
 */
export function getTokenExpiry(): number | null {
  const expiryStr = localStorage.getItem(TOKEN_EXPIRY_KEY);
  if (!expiryStr) {
    return null;
  }

  return parseInt(expiryStr, 10);
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
 * すべてのトークンをローカルストレージから削除
 */
export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(TOKEN_EXPIRY_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

/**
 * JWTトークンをパースしてペイロードを取得
 */
export function parseToken(token: string): any | null {
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
 * ローカルストレージからトークンを取得してパース
 */
export function getParsedToken(): any | null {
  const token = getToken();
  if (!token) return null;
  return parseToken(token);
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
