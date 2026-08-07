import { expect, test } from '@playwright/test';
import { stubApi } from './api-stub';
import { stubKeycloak } from './keycloak-stub';

test.describe('ログアウト', () => {
  test('Sign out で Keycloak のログアウトへ遷移し、id_token_hint を添える', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    await stubApi(page);

    await page.goto('/');
    await page.getByRole('button', { name: /linzhengen/ }).click();
    await page.getByRole('button', { name: 'Sign out' }).click();

    await expect.poll(() => keycloak.logoutRequests.length, { timeout: 15_000 }).toBe(1);

    const logoutUrl = new URL(keycloak.logoutRequests[0].url());
    // keycloak-js が idToken から自動で付与する。付いていないと Keycloak 側の
    // セッションが残り、次回ログインで確認画面が挟まる。
    expect(logoutUrl.searchParams.get('id_token_hint')).toBeTruthy();
    expect(
      logoutUrl.searchParams.get('post_logout_redirect_uri') ?? logoutUrl.searchParams.get('redirect_uri'),
    ).toContain('localhost');
  });

  test('ログアウト後にアプリへ戻ると再認証が要求される', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    expect(keycloak.authRedirects).toBe(1);

    await page.getByRole('button', { name: /linzhengen/ }).click();
    await page.getByRole('button', { name: 'Sign out' }).click();

    // ログアウト -> オリジンへ戻る -> login-required で再び認可エンドポイントへ
    await expect.poll(() => keycloak.authRedirects, { timeout: 20_000 }).toBe(2);
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
  });
});
