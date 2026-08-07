import { expect, test } from '@playwright/test';
import { stubApi } from './api-stub';
import { stubKeycloak } from './keycloak-stub';

test.describe('401 ハンドリング', () => {
  test('API が 401 を返すとセッション切れモーダルを表示する', async ({ page }) => {
    await stubKeycloak(page);
    const api = await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();

    api.failNextWithUnauthorized();
    // キャッシュされていない画面へ移動して API を叩かせる
    await page.getByRole('link', { name: 'Users', exact: true }).click();

    // antd はタイトルを本文とスクリーンリーダー用の 2 箇所に描画するので、
    // テキスト検索ではなくダイアログのアクセシブル名で特定する。
    const dialog = page.getByRole('dialog', { name: 'セッションが切れました' });
    await expect(dialog).toBeVisible({ timeout: 15_000 });
    await expect(dialog.getByText('もう一度ログインしてください。')).toBeVisible();
  });

  test('モーダルの「ログイン画面へ」でログアウトし、再認証へ進む', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    const api = await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    expect(keycloak.authRedirects).toBe(1);

    api.failNextWithUnauthorized();
    await page.getByRole('link', { name: 'Users', exact: true }).click();
    await expect(page.getByRole('dialog', { name: 'セッションが切れました' })).toBeVisible({ timeout: 15_000 });

    await page.getByRole('button', { name: 'ログイン画面へ' }).click();

    await expect.poll(() => keycloak.logoutRequests.length, { timeout: 20_000 }).toBe(1);
    await expect.poll(() => keycloak.authRedirects, { timeout: 20_000 }).toBe(2);
  });

  test('401 を受けたらメモリのトークンを破棄し、再試行に載せない', async ({ page }) => {
    const keycloak = await stubKeycloak(page);
    const api = await stubApi(page);

    await page.goto('/');
    await expect(page.getByRole('button', { name: /linzhengen/ })).toBeVisible();
    const expiredToken = keycloak.issuedAccessTokens[0];

    api.failNextWithUnauthorized();
    await page.getByRole('link', { name: 'Users', exact: true }).click();
    await expect(page.getByRole('dialog', { name: 'セッションが切れました' })).toBeVisible();

    const userRequests = () => api.requests.filter((r) => r.path === '/api/v1/users');

    // 最初のリクエストにはまだトークンが載っている
    expect(userRequests()[0].authorization).toBe(`Bearer ${expiredToken}`);

    // TanStack Query が一度リトライする。その時点で api-client は 401 を受けて
    // clearTokens() 済みなので、失効したトークンは載らない。
    await expect.poll(() => userRequests().length).toBeGreaterThan(1);
    expect(userRequests().at(-1)?.authorization).toBeUndefined();
  });
});
