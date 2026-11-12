import { http, HttpResponse, type AnyHandler } from 'msw';
import { OPERATION_TYPES } from '../api';

/** The operations list the real backend reports, so the keypad starts fully enabled in tests. */
export const OPERATIONS_BODY = {
  operations: OPERATION_TYPES.map((name) => ({
    name,
    arity: name === 'sqrt' ? 1 : 2,
    symbol: name,
    description: name,
  })),
};

// Default handlers shared by every test. Feature tests add their own with server.use(...).
export const handlers: AnyHandler[] = [
  http.get('*/api/v1/operations', () => HttpResponse.json(OPERATIONS_BODY)),
];
