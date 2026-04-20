import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

import { chunkForModule } from './src/lib/chunks';

export default defineConfig({
  plugins: [react()],
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
});
