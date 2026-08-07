import type { Page, Route } from '@playwright/test';

/**
 * バックエンド API のスタブ。認証フローの検証が目的なので、画面が描画できる
 * 最小限のレスポンスだけを返す。Authorization ヘッダーは記録して、トークンが
 * 実際に載っているかを検証できるようにする。
 */

export interface ApiStub {
  /** 受信したリクエストの (パス, Authorization ヘッダー) を到着順に記録する */
  requests: { path: string; authorization: string | undefined }[];
  /** 次回以降 /me を 401 にする */
  failNextWithUnauthorized: () => void;
}

const meResponse = {
  user: {
    id: 'user-1',
    username: 'linzhengen',
    email: 'user@example.com',
    status: 'STATUS_ACTIVE',
    groupIds: [],
  },
  groups: [],
};

export const stubApi = async (page: Page): Promise<ApiStub> => {
  let unauthorized = false;

  // index.css が Google Fonts を @import している。外部への実通信は CI で
  // 遅延・失敗の原因になるだけなので落としておく。
  await page.route(/fonts\.(googleapis|gstatic)\.com/, (route) => route.abort());

  const stub: ApiStub = {
    requests: [],
    failNextWithUnauthorized: () => {
      unauthorized = true;
    },
  };

  await page.route('**/api/v1/**', async (route: Route) => {
    const request = route.request();
    const path = new URL(request.url()).pathname;
    stub.requests.push({ path, authorization: request.headers()['authorization'] });

    if (unauthorized) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'token expired' }),
      });
      return;
    }

    const body = path.endsWith('/me/menus') ? { menus: [] } : meResponse;
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) });
  });

  return stub;
};
