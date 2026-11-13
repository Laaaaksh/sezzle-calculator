import {
  ApiError,
  CLIENT_ERROR_CODE,
  ERROR_CODE,
  NetworkError,
  TimeoutError,
  type ApiErrorCode,
} from '../../../api';
import { CLIENT_FAILURE_CODE, type CalculatorError } from './types';

const TRANSIENT_API_CODES: ReadonlySet<ApiErrorCode> = new Set<ApiErrorCode>([
  ERROR_CODE.INTERNAL,
  ERROR_CODE.SERVICE_UNAVAILABLE,
  ERROR_CODE.REQUEST_TIMEOUT,
]);

/**
 * Anything unrecognised becomes UNKNOWN, so the reducer never holds a raw Error.
 * CanceledError is not handled here: the hook swallows it first.
 */
export function toCalculatorError(error: unknown): CalculatorError {
  if (error instanceof ApiError) {
    return {
      code: error.code,
      requestId: error.requestId,
      retryable: TRANSIENT_API_CODES.has(error.code),
    };
  }
  if (error instanceof NetworkError) {
    return { code: CLIENT_FAILURE_CODE.NETWORK_ERROR, requestId: undefined, retryable: true };
  }
  if (error instanceof TimeoutError) {
    return { code: CLIENT_FAILURE_CODE.TIMEOUT, requestId: undefined, retryable: true };
  }
  return { code: CLIENT_ERROR_CODE.UNKNOWN, requestId: undefined, retryable: false };
}
