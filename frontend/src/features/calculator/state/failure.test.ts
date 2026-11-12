import { describe, expect, it } from 'vitest';
import { ApiError, CanceledError, NetworkError, TimeoutError } from '../../../api';
import { toCalculatorError } from './failure';

describe('toCalculatorError', () => {
  it('keeps the code, message and request ID of an API error', () => {
    const error = new ApiError({
      code: 'DIVISION_BY_ZERO',
      message: 'Cannot divide by zero.',
      status: 422,
      requestId: 'req-9',
    });

    expect(toCalculatorError(error)).toEqual({
      code: 'DIVISION_BY_ZERO',
      message: 'Cannot divide by zero.',
      requestId: 'req-9',
      retryable: false,
    });
  });

  it('marks a network error as retryable', () => {
    expect(toCalculatorError(new NetworkError())).toMatchObject({
      code: 'NETWORK_ERROR',
      retryable: true,
    });
  });

  it('marks a timeout as retryable', () => {
    expect(toCalculatorError(new TimeoutError(8000))).toMatchObject({
      code: 'TIMEOUT',
      retryable: true,
    });
  });

  it.each([
    ['a canceled request', new CanceledError()],
    ['a plain Error', new Error('boom')],
    ['a thrown string', 'boom'],
  ])('maps %s to UNKNOWN', (_label, thrown) => {
    expect(toCalculatorError(thrown)).toMatchObject({ code: 'UNKNOWN', retryable: false });
  });
});
