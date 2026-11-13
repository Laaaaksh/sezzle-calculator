import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';

const DEFAULT_DEV_API_TARGET = 'http://localhost:8080';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const configured = env.VITE_DEV_API_TARGET;
  // An unset or empty variable falls back to the default so the proxy never has a blank target.
  const target =
    configured !== undefined && configured !== '' ? configured : DEFAULT_DEV_API_TARGET;

  return {
    plugins: [react()],
    server: {
      // types.test.ts reads ../docs/API.md to check the API types against the contract.
      // Only that folder is opened up. A single-file entry is not enough: Vite then denies the
      // file's ?raw import, so the narrowest working scope is the docs folder itself.
      fs: { allow: ['.', '../docs'] },
      // Proxying keeps the browser on one origin in dev, so there are no CORS issues.
      proxy: {
        '/api': { target, changeOrigin: true },
        '/health': { target, changeOrigin: true },
      },
    },
    test: {
      environment: 'jsdom',
      // Vitest blanks CSS by default, even with ?raw. The token and breakpoint tests read the real
      // files to check contrast ratios and breakpoints.
      // Only ?raw imports are opened up. Applying stylesheets to every component test would make
      // jsdom hide the history button, because it ignores media queries and treats them as true.
      css: { include: [/\.css\?raw/] },
      setupFiles: ['src/test/setup.ts'],
      coverage: {
        provider: 'v8',
        reporter: ['text', 'html', 'lcov', 'json-summary'],
        reportsDirectory: 'coverage',
        // Only source files: .gitkeep and .css would otherwise count as 0% and drag the totals down.
        include: ['src/**/*.{ts,tsx}'],
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
