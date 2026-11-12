import { ApiError, NetworkError, TimeoutError } from '../../../api';
import type { CalculatorError } from './types';

const NETWORK_MESSAGE = 'Could not reach the server.';
const TIMEOUT_MESSAGE = 'The request took too long.';
const UNKNOWN_MESSAGE = 'Something went wrong.';

/** Server-side failures that may clear up on their own, so sending the same request again can work. */
const TRANSIENT_API_CODES: ReadonlySet<string> = new Set(['INTERNAL', 'SERVICE_UNAVAILABLE', 'REQUEST_TIMEOUT']);

/**
 * Turns whatever the API client threw into the one error shape the reducer stores.
 * Anything unrecognised becomes UNKNOWN, so the reducer never holds a raw Error.
 * CanceledError is not handled here on purpose: the hook swallows it before it gets this far.
 */
export function toCalculatorError(error: unknown): CalculatorError {
  if (error instanceof ApiError) {
    return {
      code: error.code,
      message: error.message,
      requestId: error.requestId,
      retryable: TRANSIENT_API_CODES.has(error.code),
    };
  }
  if (error instanceof NetworkError) {
    return {
      code: 'NETWORK_ERROR',
      message: NETWORK_MESSAGE,
      requestId: undefined,
      retryable: true,
    };
  }
  if (error instanceof TimeoutError) {
    return { code: 'TIMEOUT', message: TIMEOUT_MESSAGE, requestId: undefined, retryable: true };
  }
  return { code: 'UNKNOWN', message: UNKNOWN_MESSAGE, requestId: undefined, retryable: false };
}
