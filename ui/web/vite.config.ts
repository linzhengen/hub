import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import svgr from "vite-plugin-svgr";
import * as path from 'path';
import {defineConfig} from 'vite';

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
