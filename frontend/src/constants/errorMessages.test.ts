import { describe, expect, it } from 'vitest';
import { ERROR_CODES } from '../api';
import type { CalculatorError } from '../features/calculator/state';
import { ERROR_MESSAGES, GENERIC_ERROR_MESSAGE, messageForError } from './errorMessages';

function errorWith(code: string, requestId?: string): CalculatorError {
  // The cast stands in for a code from a newer server that slipped past the types.
  return { code: code as CalculatorError['code'], message: 'raw', requestId, retryable: false };
}

describe('messageForError', () => {
  it.each(ERROR_CODES)('has friendly copy for %s', (code) => {
    const message = messageForError(errorWith(code));

    expect(message).toBe(ERROR_MESSAGES[code]);
    expect(message).not.toBe(GENERIC_ERROR_MESSAGE);
  });

  it('says you cannot divide by zero', () => {
    expect(messageForError(errorWith('DIVISION_BY_ZERO'))).toBe("You can't divide by zero.");
  });

  it.each(['NETWORK_ERROR', 'TIMEOUT', 'INVALID_RESPONSE', 'RESULT_TOO_LONG_TO_REUSE'])(
    'has copy for %s',
    (code) => {
      expect(messageForError(errorWith(code))).not.toBe(GENERIC_ERROR_MESSAGE);
    },
  );

  it('falls back to a generic message with the request ID for UNKNOWN', () => {
    expect(messageForError(errorWith('UNKNOWN', 'req-77'))).toBe(
      `${GENERIC_ERROR_MESSAGE} Request ID: req-77`,
    );
  });

  it('falls back to the generic message for a code with no copy', () => {
    expect(messageForError(errorWith('BRAND_NEW_CODE', 'req-8'))).toContain('req-8');
  });

  it('does not treat an Object method name as a known code', () => {
    expect(messageForError(errorWith('toString'))).toBe(GENERIC_ERROR_MESSAGE);
  });

  it('gives the plain generic message when there is no request ID', () => {
    expect(messageForError(errorWith('UNKNOWN'))).toBe(GENERIC_ERROR_MESSAGE);
  });
});
