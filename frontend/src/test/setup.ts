import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll, vi } from 'vitest';
import { server } from './server';

// 'error' makes any request without a handler fail the test, so no test silently hits the network.
// (MSW 3 named this option onUnhandledFrame; older versions call it onUnhandledRequest.)
beforeAll(() => {
  server.listen({ onUnhandledFrame: 'error' });
  // axe-core probes canvas support, and jsdom answers every probe with a "not implemented" error
  // on the console. A null context is the documented "no canvas" answer, so axe moves on quietly.
  // Only getContext is replaced: any other jsdom error still prints.
  vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(() => null);
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
});

afterAll(() => {
  server.close();
});
