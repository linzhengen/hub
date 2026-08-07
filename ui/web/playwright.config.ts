import { defineConfig, devices } from '@playwright/test';

const PORT = 3100;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  // dev server はルートごとに初回だけオンデマンドでコンパイルする。#68 で
  // ページを遅延読み込みにしたぶん初回遷移が重いので、並列度は抑えめにする。
  workers: 2,
  timeout: 60_000,
  expect: { timeout: 15_000 },
  reporter: process.env.CI
    ? ([['github'], ['list'], ['html', { open: 'never' }]] as const)
    : 'list',
  use: {
    baseURL: `http://localhost:${PORT}`,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // 既に Chromium が入っている環境（開発コンテナ等）では、Playwright の
        // バージョンに紐づくビルドを取り直さずにそれを使う。
        ...(process.env.PLAYWRIGHT_CHROMIUM_PATH
          ? { launchOptions: { executablePath: process.env.PLAYWRIGHT_CHROMIUM_PATH } }
          : {}),
      },
    },
  ],
  webServer: {
    // あえて preview ではなく dev server を使う。StrictMode が effect を二重に
    // 実行するのは開発ビルドだけで、Keycloak の初期化はまさにそこで壊れていた。
    // 本番ビルドで動かすとその条件を踏まなくなる。HMR は不要なので切っておく。
    command: `pnpm vite --port ${PORT} --strictPort`,
    url: `http://localhost:${PORT}`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: { DISABLE_HMR: 'true' },
  },
});
