import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

const apiBaseUrl = process.env.TEST_API_BASE_URL ?? 'http://127.0.0.1:18080';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  define: {
    'import.meta.env.VITE_API_MODE': JSON.stringify('http'),
    'import.meta.env.VITE_API_BASE_URL': JSON.stringify(apiBaseUrl),
    'import.meta.env.PROD': 'false',
  },
  test: {
    include: ['tests/integration/**/*.test.{ts,tsx}'],
    exclude: ['**/node_modules/**', '**/dist/**'],
    setupFiles: ['./src/test/setup.ts', './tests/integration/setup.ts'],
    testTimeout: 30000,
    hookTimeout: 30000,
    env: {
      TEST_API_BASE_URL: apiBaseUrl,
    },
  },
});
