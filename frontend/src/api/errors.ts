import { CANCELED_MESSAGE, ERROR_MESSAGES } from '../constants/errorMessages';
import type { ApiErrorCode, ErrorFields } from './types';

export interface ApiErrorInit {
  code: ApiErrorCode;
  message: string;
  status: number;
  requestId?: string | undefined;
  fields?: ErrorFields | undefined;
}

/** The server answered with an error, or with a body that breaks the contract. */
export class ApiError extends Error {
  override readonly name = 'ApiError';
  readonly code: ApiErrorCode;
  readonly status: number;
  /** Quote this to support. Comes from the error body or the X-Request-ID header. */
  readonly requestId: string | undefined;
  readonly fields: ErrorFields | undefined;

  constructor(init: ApiErrorInit, options?: ErrorOptions) {
    super(init.message, options);
    this.code = init.code;
    this.status = init.status;
    this.requestId = init.requestId;
    this.fields = init.fields;
  }
}

/** No response arrived: the server is down, the connection dropped, or the browser is offline. */
export class NetworkError extends Error {
  override readonly name = 'NetworkError';

  constructor(options?: ErrorOptions) {
    super(ERROR_MESSAGES.NETWORK_ERROR, options);
  }
}

/** The request ran longer than its time limit and was cancelled by the client. */
export class TimeoutError extends Error {
  override readonly name = 'TimeoutError';
  readonly timeoutMs: number;

  constructor(timeoutMs: number, options?: ErrorOptions) {
    super(ERROR_MESSAGES.TIMEOUT, options);
    this.timeoutMs = timeoutMs;
  }
}

/**
 * The caller aborted the request with its own AbortSignal. A typed error (not a DOMException)
 * so a UI can tell "I cancelled this" apart from a failure with one instanceof check.
 */
export class CanceledError extends Error {
  override readonly name = 'CanceledError';

  constructor(options?: ErrorOptions) {
    super(CANCELED_MESSAGE, options);
  }
}
