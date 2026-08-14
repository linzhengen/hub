import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';
import { stubApi } from './api-stub';
import { stubKeycloak } from './keycloak-stub';

/**
 * Chat の API スタブ。`stubApi` の後に登録することで、`**\/api/v1/**` の
 * 総括ルートより優先される（Playwright は後から登録したルートを先に照合する）。
 *
 * SendMessage は server-streaming なので、grpc-gateway と同じく
 * 改行区切りの `{"result": ...}` を返す。
 */
const stubChat = async (page: Page) => {
  const sessions = [{ id: 's1', title: 'First chat', userId: 'user-1' }];
  const messages: { id: string; role: string; content: string }[] = [];

  await page.route('**/api/v1/chat/sessions/*/messages', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ messages }),
      });
      return;
    }

    const content = (route.request().postDataJSON() as { content: string }).content;
    messages.push({ id: `m${messages.length + 1}`, role: 'user', content });
    messages.push({ id: `m${messages.length + 1}`, role: 'assistant', content: 'Hello from the mock backend.' });

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body:
        [
          JSON.stringify({ result: { delta: 'Hello from the ' } }),
          JSON.stringify({ result: { delta: 'mock backend.' } }),
          JSON.stringify({ result: { done: true } }),
        ].join('\n') + '\n',
    });
  });

  await page.route('**/api/v1/chat/sessions', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ sessions }),
    });
  });
};

test.describe('AI chat', () => {
  test('メッセージを送ると応答が会話に表示される', async ({ page }) => {
    await stubKeycloak(page);
    await stubApi(page);
    await stubChat(page);

    await page.goto('/chat');

    await expect(page.getByText('First chat')).toBeVisible();

    await page.getByLabel('Message').fill('こんにちは');
    await page.getByRole('button', { name: 'Send' }).click();

    await expect(page.getByText('こんにちは')).toBeVisible();
    await expect(page.getByText('Hello from the mock backend.')).toBeVisible();
  });
});
