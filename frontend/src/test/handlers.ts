import type { AnyHandler } from 'msw';

// Default handlers shared by every test. Feature tests add their own with server.use(...).
export const handlers: AnyHandler[] = [];
