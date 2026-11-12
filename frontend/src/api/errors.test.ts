import { describe, expect, it } from 'vitest';
import { ApiError, CanceledError, NetworkError, TimeoutError } from './errors';

describe('typed errors', () => {
  it('ApiError carries code, message, status, request ID and fields', () => {
    const error = new ApiError({
      code: 'INVALID_OPERAND',
      message: 'An operand is not a valid number.',
      status: 422,
      requestId: 'req-1',
      fields: { operand_index: 1, reason: 'exponent' },
    });

    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe('ApiError');
    expect(error.code).toBe('INVALID_OPERAND');
    expect(error.message).toBe('An operand is not a valid number.');
    expect(error.status).toBe(422);
    expect(error.requestId).toBe('req-1');
    expect(error.fields).toEqual({ operand_index: 1, reason: 'exponent' });
  });

  it('ApiError keeps the cause', () => {
    const cause = new Error('root');

    const error = new ApiError({ code: 'UNKNOWN', message: 'x', status: 500 }, { cause });

    expect(error.cause).toBe(cause);
    expect(error.requestId).toBeUndefined();
  });

  it('NetworkError, TimeoutError and CanceledError set name and keep the cause', () => {
    const cause = new TypeError('fetch failed');

    const network = new NetworkError({ cause });
    const timeout = new TimeoutError(250, { cause });
    const canceled = new CanceledError({ cause });

    expect(network.name).toBe('NetworkError');
    expect(timeout.name).toBe('TimeoutError');
    expect(timeout.timeoutMs).toBe(250);
    expect(canceled.name).toBe('CanceledError');
    expect([network.cause, timeout.cause, canceled.cause]).toEqual([cause, cause, cause]);
  });
});
