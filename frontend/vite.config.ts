import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';

const DEFAULT_DEV_API_TARGET = 'http://localhost:8080';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const configured = env.VITE_DEV_API_TARGET;
  // An unset or empty variable falls back to the default so the proxy never has a blank target.
  const target =
    configured !== undefined && configured !== '' ? configured : DEFAULT_DEV_API_TARGET;

  return {
    plugins: [react()],
    server: {
      // types.test.ts imports ../docs/API.md to check the API types against the contract.
      // Only that folder is opened up; the rest of the repo stays outside the dev server.
      fs: { allow: ['.', '../docs'] },
      // Proxying keeps the browser on one origin in dev, so there are no CORS issues.
      proxy: {
        '/api': { target, changeOrigin: true },
        '/health': { target, changeOrigin: true },
      },
    },
    test: {
      environment: 'jsdom',
      setupFiles: ['src/test/setup.ts'],
      coverage: {
        provider: 'v8',
        reporter: ['text', 'html', 'lcov', 'json-summary'],
        reportsDirectory: 'coverage',
        include: ['src/**'],
        exclude: [
          'src/main.tsx',
          'src/**/*.test.{ts,tsx}',
          'src/test/**',
          'src/**/*.d.ts',
          'src/**/types.ts',
        ],
        thresholds: { lines: 80, functions: 80, branches: 80, statements: 80 },
      },
    },
  };
});
