import {
  ApiError,
  CLIENT_ERROR_CODE,
  ERROR_CODE,
  NetworkError,
  TimeoutError,
  type ApiErrorCode,
} from '../../../api';
import { CLIENT_FAILURE_CODE, type CalculatorError } from './types';

/** Server-side failures that may clear up on their own, so sending the same request again can work. */
const TRANSIENT_API_CODES: ReadonlySet<ApiErrorCode> = new Set<ApiErrorCode>([
  ERROR_CODE.INTERNAL,
  ERROR_CODE.SERVICE_UNAVAILABLE,
  ERROR_CODE.REQUEST_TIMEOUT,
]);

/**
 * Turns whatever the API client threw into the one error shape the reducer stores.
 * Anything unrecognised becomes UNKNOWN, so the reducer never holds a raw Error.
 * CanceledError is not handled here on purpose: the hook swallows it before it gets this far.
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
