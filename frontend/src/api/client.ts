import {
  API_PATHS,
  DEFAULT_API_BASE_URL,
  DEFAULT_REQUEST_TIMEOUT_MS,
  HEADER_NAMES,
  HTTP_METHODS,
  JSON_MEDIA_TYPE,
} from '../constants/api';
import { ApiError, CanceledError, NetworkError, TimeoutError } from './errors';
import { isCalculateResponse, isErrorBodyShape, isErrorCode, isOperationsResponse } from './guards';
import type { CalculateRequest, CalculateResponse, OperationInfo } from './types';

export interface RequestOptions {
  /** Aborts the request. The call then rejects with a CanceledError. */
  signal?: AbortSignal | undefined;
  /** Overrides the client's time limit for this one call. */
  timeoutMs?: number | undefined;
}

export interface ApiClientConfig {
  /** Prefix for every path. Empty means same origin. A trailing slash is ignored. */
  baseUrl?: string | undefined;
  /** Default time limit per request in milliseconds. */
  timeoutMs?: number | undefined;
}

export interface ApiClient {
  calculate: (request: CalculateRequest, options?: RequestOptions) => Promise<CalculateResponse>;
  getOperations: (options?: RequestOptions) => Promise<OperationInfo[]>;
}

const INVALID_RESPONSE_MESSAGE = 'The server sent a response this app cannot read.';
const UNKNOWN_ERROR_MESSAGE = 'The server returned an unexpected error.';

/** Turns the raw env value into a base URL: unset or blank becomes the default, no trailing slash. */
export function resolveBaseUrl(raw: string | undefined): string {
  const trimmed = (raw ?? '').trim();
  return (trimmed === '' ? DEFAULT_API_BASE_URL : trimmed).replace(/\/+$/, '');
}

function assertValidTimeout(timeoutMs: number): void {
  // A zero, negative or NaN limit would cancel every request at once, so fail loudly instead.
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) {
    throw new RangeError(`timeoutMs must be a positive number, got ${timeoutMs}`);
  }
}

function parseJson(text: string): unknown {
  try {
    const parsed: unknown = JSON.parse(text);
    return parsed;
  } catch {
    return undefined;
  }
}

function buildApiError(response: Response, text: string): ApiError {
  const body = parseJson(text);
  const headerRequestId = response.headers.get(HEADER_NAMES.requestId) ?? undefined;

  if (!isErrorBodyShape(body)) {
    return new ApiError({
      code: 'UNKNOWN',
      message: UNKNOWN_ERROR_MESSAGE,
      status: response.status,
      requestId: headerRequestId,
    });
  }

  const { code, message, requestId, fields } = body.error;
  return new ApiError({
    // A code from a newer server is kept out of the typed union and reported as UNKNOWN.
    code: isErrorCode(code) ? code : 'UNKNOWN',
    message,
    status: response.status,
    requestId,
    fields,
  });
}

function buildInvalidResponseError(response: Response): ApiError {
  return new ApiError({
    code: 'INVALID_RESPONSE',
    message: INVALID_RESPONSE_MESSAGE,
    status: response.status,
    requestId: response.headers.get(HEADER_NAMES.requestId) ?? undefined,
  });
}

/** Creates a client. Tests pass their own config; the app uses the default client below. */
export function createApiClient(config: ApiClientConfig = {}): ApiClient {
  const baseUrl = resolveBaseUrl(config.baseUrl);
  const defaultTimeoutMs = config.timeoutMs ?? DEFAULT_REQUEST_TIMEOUT_MS;
  assertValidTimeout(defaultTimeoutMs);

  async function send<T>(
    path: string,
    init: { method: string; body?: string },
    isExpected: (value: unknown) => value is T,
    options: RequestOptions,
  ): Promise<T> {
    const timeoutMs = options.timeoutMs ?? defaultTimeoutMs;
    assertValidTimeout(timeoutMs);
    const callerSignal = options.signal;
    if (callerSignal?.aborted) {
      throw new CanceledError({ cause: callerSignal.reason });
    }

    // One controller is aborted by either the caller or the timer. The flag records which one,
    // because fetch rejects with the same kind of error in both cases.
    const controller = new AbortController();
    const state = { timedOut: false };
    const timer = setTimeout(() => {
      state.timedOut = true;
      controller.abort();
    }, timeoutMs);
    const onCallerAbort = () => {
      controller.abort(callerSignal?.reason);
    };
    callerSignal?.addEventListener('abort', onCallerAbort, { once: true });

    let response: Response;
    let text: string;
    try {
      response = await fetch(`${baseUrl}${path}`, {
        ...init,
        headers: {
          [HEADER_NAMES.accept]: JSON_MEDIA_TYPE,
          ...(init.body === undefined ? {} : { [HEADER_NAMES.contentType]: JSON_MEDIA_TYPE }),
        },
        signal: controller.signal,
      });
      // The timer keeps running while the body downloads, so a stalled body also times out.
      text = await response.text();
    } catch (cause) {
      if (state.timedOut) {
        throw new TimeoutError(timeoutMs, { cause });
      }
      if (callerSignal?.aborted) {
        throw new CanceledError({ cause });
      }
      throw new NetworkError({ cause });
    } finally {
      clearTimeout(timer);
      callerSignal?.removeEventListener('abort', onCallerAbort);
    }

    if (!response.ok) {
      throw buildApiError(response, text);
    }
    const body = parseJson(text);
    if (!isExpected(body)) {
      throw buildInvalidResponseError(response);
    }
    return body;
  }

  return {
    calculate: (request, options = {}) =>
      send(
        API_PATHS.calculate,
        { method: HTTP_METHODS.post, body: JSON.stringify(request) },
        isCalculateResponse,
        options,
      ),
    getOperations: async (options = {}) => {
      const body = await send(
        API_PATHS.operations,
        { method: HTTP_METHODS.get },
        isOperationsResponse,
        options,
      );
      return body.operations;
    },
  };
}

const defaultClient = createApiClient({
  baseUrl: resolveBaseUrl(import.meta.env.VITE_API_BASE_URL),
});

/** Runs one operation on the server. Rejects with ApiError, NetworkError, TimeoutError or CanceledError. */
export const calculate = defaultClient.calculate;

/** Lists the operations the server supports. Rejects like calculate. */
export const getOperations = defaultClient.getOperations;
