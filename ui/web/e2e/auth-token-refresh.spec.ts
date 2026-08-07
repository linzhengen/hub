import { expect, test } from '@playwright/test';
import { stubApi } from './api-stub';
import { stubKeycloak } from './keycloak-stub';

test.describe('トークンリフレッシュ', () => {
  test('期限切れでリフレッシュし、以降の API 呼び出しは新しいトークンを使う', async ({ page }) => {
    // keycloak-js は expires_in をもとに onTokenExpired のタイマーを張るので、
    // 寿命を短くしておけばテスト中に自動リフレッシュが走る。
    const keycloak = await stubKeycloak(page, { accessTokenLifetimeSeconds: 3 });
    const api = await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();

    const firstToken = keycloak.issuedAccessTokens[0];
    expect(keycloak.tokenGrants).toEqual(['authorization_code']);

    // 次に発行するトークンは長寿命にして、リフレッシュを一度だけに収める
    keycloak.setAccessTokenLifetime(300);
    await expect.poll(() => keycloak.tokenGrants, { timeout: 20_000 }).toContain('refresh_token');

    const refreshedToken = keycloak.issuedAccessTokens[0];
    expect(refreshedToken).not.toBe(firstToken);

    // リロードせずにクライアント側遷移する。ページを読み直すと認可コードから
    // 取り直してしまい、リフレッシュ後のトークンが使われたか確認できない。
    await page.getByRole('link', { name: 'Users', exact: true }).click();
    await expect.poll(() => api.requests.some((r) => r.path === '/api/v1/users'), { timeout: 15_000 }).toBe(true);

    const usersRequest = api.requests.find((r) => r.path === '/api/v1/users');
    expect(usersRequest?.authorization).toBe(`Bearer ${refreshedToken}`);
    expect(usersRequest?.authorization).not.toBe(`Bearer ${firstToken}`);
  });

  test('リフレッシュが失敗したらログイン画面へ送り直す', async ({ page }) => {
    const keycloak = await stubKeycloak(page, { accessTokenLifetimeSeconds: 3 });
    await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    expect(keycloak.authRedirects).toBe(1);

    // 以降の /token をすべて失敗させる
    await page.route(`${(await import('./keycloak-stub')).REALM_URL}/protocol/openid-connect/token`, (route) =>
      route.fulfill({ status: 400, contentType: 'application/json', body: '{"error":"invalid_grant"}' }),
    );

    // リフレッシュ失敗 -> keycloak.login() で認可エンドポイントへ
    await expect.poll(() => keycloak.authRedirects, { timeout: 20_000 }).toBeGreaterThan(1);
  });
});
