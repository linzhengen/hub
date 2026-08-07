import { expect, test } from '@playwright/test';
import { stubApi } from './api-stub';
import { stubKeycloak } from './keycloak-stub';

test.describe('ログイン', () => {
  test('未認証なら認可エンドポイントへ送られ、コールバックで認証が完了する', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    const api = await stubApi(page);

    await page.goto('/');

    // 認可エンドポイント経由で戻り、保護されたページが描画される
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    expect(keycloak.authRedirects).toBe(1);
    expect(keycloak.tokenGrants).toEqual(['authorization_code']);
    expect(api.requests.length).toBeGreaterThan(0);
  });

  test('取得したアクセストークンを API の Authorization ヘッダーに載せる', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    const api = await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();

    const meRequest = api.requests.find((r) => r.path === '/api/v1/me');
    expect(meRequest?.authorization).toBe(`Bearer ${keycloak.issuedAccessTokens[0]}`);
  });

  test('リロードのたびに SSO セッションから再認証する（トークンは永続化しない）', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();

    // トークンはメモリのみ。localStorage / sessionStorage に漏れていないこと
    const persisted = await page.evaluate(() => ({
      local: JSON.stringify(window.localStorage),
      session: JSON.stringify(window.sessionStorage),
    }));
    expect(persisted.local).not.toContain('Bearer');
    expect(persisted.local).not.toContain('eyJhbGci');
    expect(persisted.session).not.toContain('eyJhbGci');

    await page.reload();
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    expect(keycloak.authRedirects).toBe(2);
  });
});
