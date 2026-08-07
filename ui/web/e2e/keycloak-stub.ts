import type { Page, Request, Route } from '@playwright/test';

/**
 * 実 Keycloak を立てずに認証フローを通すためのスタブ。
 *
 * アプリ側（keycloak-js / AuthProvider / api-client）は一切差し替えず、
 * ネットワーク境界だけを Playwright の route interception で置き換える。
 * したがって検証できるのは「アプリが OIDC の各ステップをどう扱うか」であって、
 * Keycloak 自身の挙動（署名検証、セッション管理、SMTP 等）ではない。
 */

// vite.config.ts の dev server と同じ既定値。VITE_KEYCLOAK_* は未設定で動かす。
export const KEYCLOAK_URL = 'http://localhost:8080';
export const REALM = 'hub';
export const REALM_URL = `${KEYCLOAK_URL}/realms/${REALM}`;

const base64url = (value: object): string =>
  Buffer.from(JSON.stringify(value)).toString('base64url');

export interface TokenClaims {
  sub?: string;
  preferred_username?: string;
  email?: string;
  email_verified?: boolean;
  realm_access?: { roles: string[] };
  nonce?: string;
}

/**
 * 署名なしの JWT を組み立てる。keycloak-js はクライアント側で署名検証を
 * しないため、ペイロードが読めれば十分。
 */
export const makeJwt = (claims: TokenClaims, expiresInSeconds: number): string => {
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    sub: 'user-1',
    preferred_username: 'linzhengen',
    email: 'user@example.com',
    email_verified: true,
    realm_access: { roles: ['user'] },
    ...claims,
    iat: now,
    exp: now + expiresInSeconds,
    iss: REALM_URL,
    typ: 'Bearer',
    azp: 'hub-web',
  };

  return `${base64url({ alg: 'RS256', typ: 'JWT' })}.${base64url(payload)}.signature`;
};

export interface KeycloakStub {
  /** /token に来た grant_type を到着順に記録する */
  tokenGrants: string[];
  /** 認可エンドポイントへ送られた回数（= ログイン画面に飛ばされた回数） */
  authRedirects: number;
  /** ログアウトエンドポイントへ送られたリクエスト */
  logoutRequests: Request[];
  /** 次に /token が返すアクセストークンの寿命（秒） */
  setAccessTokenLifetime: (seconds: number) => void;
  /** 次に /token が返すクレーム */
  setClaims: (claims: TokenClaims) => void;
  /** 発行済みアクセストークンを新しい順に保持する */
  issuedAccessTokens: string[];
}

export interface StubOptions {
  /** 初回に発行するアクセストークンの寿命。短くすると自動リフレッシュが走る */
  accessTokenLifetimeSeconds?: number;
  claims?: TokenClaims;
}

/**
 * Keycloak の認可 / トークン / ログアウトエンドポイントをスタブする。
 */
export const stubKeycloak = async (page: Page, options: StubOptions = {}): Promise<KeycloakStub> => {
  const stub: KeycloakStub = {
    tokenGrants: [],
    authRedirects: 0,
    logoutRequests: [],
    issuedAccessTokens: [],
    setAccessTokenLifetime: (seconds) => {
      lifetime = seconds;
    },
    setClaims: (next) => {
      claims = next;
    },
  };

  let lifetime = options.accessTokenLifetimeSeconds ?? 300;
  let claims: TokenClaims = options.claims ?? {};
  // keycloak-js は id_token の nonce を認可リクエストで送った値と突き合わせる
  let nonce: string | undefined;

  // 認可エンドポイント: ログイン画面を出す代わりに、そのまま redirect_uri へ
  // code と state を付けて返す。state は keycloak-js が保存した値をそのまま
  // 返さないとコールバックが破棄されるため、リクエストから読み取って echo する。
  await page.route(`${REALM_URL}/protocol/openid-connect/auth*`, async (route: Route) => {
    stub.authRedirects += 1;

    const requested = new URL(route.request().url());
    const redirectUri = requested.searchParams.get('redirect_uri') ?? '';
    const state = requested.searchParams.get('state') ?? '';
    nonce = requested.searchParams.get('nonce') ?? undefined;

    const params = new URLSearchParams({
      state,
      session_state: 'session-1',
      iss: REALM_URL,
      code: 'authorization-code',
    });

    // keycloak-js の既定の response_mode は fragment。クエリ文字列で返すと
    // コールバックとして認識されず、ログイン画面へのループになる。
    const callback = new URL(redirectUri);
    if (requested.searchParams.get('response_mode') === 'query') {
      params.forEach((value, key) => callback.searchParams.set(key, value));
    } else {
      callback.hash = params.toString();
    }

    await route.fulfill({ status: 302, headers: { location: callback.toString() } });
  });

  await page.route(`${REALM_URL}/protocol/openid-connect/token`, async (route: Route) => {
    const body = new URLSearchParams(route.request().postData() ?? '');
    stub.tokenGrants.push(body.get('grant_type') ?? 'unknown');

    const accessToken = makeJwt(claims, lifetime);
    stub.issuedAccessTokens.unshift(accessToken);

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        access_token: accessToken,
        refresh_token: makeJwt(claims, 1800),
        id_token: makeJwt({ ...claims, nonce }, lifetime),
        token_type: 'Bearer',
        expires_in: lifetime,
        refresh_expires_in: 1800,
        session_state: 'session-1',
        scope: 'openid profile email',
      }),
    });
  });

  // ログアウトはブラウザ遷移。アプリのオリジンへ戻す。
  await page.route(`${REALM_URL}/protocol/openid-connect/logout*`, async (route: Route) => {
    stub.logoutRequests.push(route.request());

    const requested = new URL(route.request().url());
    const redirectUri = requested.searchParams.get('post_logout_redirect_uri') ?? requested.searchParams.get('redirect_uri');

    await route.fulfill({ status: 302, headers: { location: redirectUri ?? '/' } });
  });

  return stub;
};
