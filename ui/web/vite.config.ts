import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import svgr from "vite-plugin-svgr";
import * as path from 'path';
import {defineConfig} from 'vitest/config';

export default defineConfig(() => {
  return {
    plugins: [
      react(),
      svgr({
        svgrOptions: {
          icon: true,
          // This will transform your SVG to a React component
          exportType: "default",
        },
      }),
      tailwindcss(),
    ],
    build: {
      emptyOutDir: false,
    },
    test: {
      environment: 'jsdom',
      include: ['src/**/*.{test,spec}.{ts,tsx}'],
      setupFiles: ['./src/test/setup.ts'],
      // restoreMocks は spyOn のみが対象なので、vi.mock で作った vi.fn() の
      // 呼び出し履歴がテスト間で漏れないよう clearMocks も有効にする
      clearMocks: true,
      restoreMocks: true,
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      // HMR is disabled in AI Studio via DISABLE_HMR env var.
      // Do not modifyâfile watching is disabled to prevent flickering during agent edits.
      hmr: process.env.DISABLE_HMR !== 'true',
      proxy: {
        '/api': {
          target: 'http://localhost:9090',
          changeOrigin: true,
        },
      },
    },
  };
});
