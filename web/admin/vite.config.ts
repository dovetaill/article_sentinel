import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';

import { chunkForModule } from './src/lib/chunks';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '');
  const adminApiBaseUrl = env.ADMIN_API_BASE_URL || 'http://127.0.0.1:8080';

  return {
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/auth': {
          target: adminApiBaseUrl,
          changeOrigin: true
        },
        '/api': {
          target: adminApiBaseUrl,
          changeOrigin: true
        }
      }
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks: chunkForModule
        }
      }
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: './src/test/setup.ts',
      css: true
    }
  };
});
