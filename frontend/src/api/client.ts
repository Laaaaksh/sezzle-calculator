import {
  API_PATHS,
  DEFAULT_API_BASE_URL,
  DEFAULT_REQUEST_TIMEOUT_MS,
  HEADER_NAMES,
  HTTP_METHODS,
  JSON_MEDIA_TYPE,
} from '../constants/api';
import { ERROR_MESSAGES, GENERIC_ERROR_MESSAGE } from '../constants/errorMessages';
import { ApiError, CanceledError, NetworkError, TimeoutError } from './errors';
import { isCalculateResponse, isErrorBodyShape, isErrorCode, isOperationsResponse } from './guards';
import {
  CLIENT_ERROR_CODE,
  type CalculateRequest,
  type CalculateResponse,
  type OperationsResponse,
} from './types';

export interface RequestOptions {
  signal?: AbortSignal | undefined;
  timeoutMs?: number | undefined;
}

export interface ApiClientConfig {
  /** Prefix for every path. Empty means same origin. A trailing slash is ignored. */
  baseUrl?: string | undefined;
  timeoutMs?: number | undefined;
}

export interface ApiClient {
  calculate: (request: CalculateRequest, options?: RequestOptions) => Promise<CalculateResponse>;
  getOperations: (options?: RequestOptions) => Promise<OperationsResponse>;
}

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
      code: CLIENT_ERROR_CODE.UNKNOWN,
      message: GENERIC_ERROR_MESSAGE,
      status: response.status,
      requestId: headerRequestId,
    });
  }

  const { code, message, requestId, fields } = body.error;
  return new ApiError({
    // A code from a newer server is kept out of the typed union and reported as UNKNOWN.
    code: isErrorCode(code) ? code : CLIENT_ERROR_CODE.UNKNOWN,
    message,
    status: response.status,
    requestId,
    fields,
  });
}

function buildInvalidResponseError(response: Response): ApiError {
  return new ApiError({
    code: CLIENT_ERROR_CODE.INVALID_RESPONSE,
    message: ERROR_MESSAGES.INVALID_RESPONSE,
    status: response.status,
    requestId: response.headers.get(HEADER_NAMES.requestId) ?? undefined,
  });
}

interface RequestSpec<T> {
  path: string;
  method: string;
  body?: string | undefined;
  isExpected: (value: unknown) => value is T;
}

/**
 * One abort signal fed by two sources: the caller's signal and a timer. `timedOut` records which
 * one fired, because fetch rejects with the same kind of error in both cases.
 */
interface AbortScope {
  signal: AbortSignal;
  timedOut: () => boolean;
  /** Always call this (in a finally): it stops the timer and unhooks the caller's signal. */
  dispose: () => void;
}

function createAbortScope(callerSignal: AbortSignal | undefined, timeoutMs: number): AbortScope {
  const controller = new AbortController();
  let timedOut = false;
  const timer = setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeoutMs);
  const onCallerAbort = () => {
    controller.abort(callerSignal?.reason);
  };
  callerSignal?.addEventListener('abort', onCallerAbort, { once: true });

  return {
    signal: controller.signal,
    timedOut: () => timedOut,
    dispose: () => {
      clearTimeout(timer);
      callerSignal?.removeEventListener('abort', onCallerAbort);
    },
  };
}

async function fetchText(
  url: string,
  spec: Pick<RequestSpec<unknown>, 'method' | 'body'>,
  signal: AbortSignal,
): Promise<{ response: Response; text: string }> {
  const response = await fetch(url, {
    method: spec.method,
    ...(spec.body === undefined ? {} : { body: spec.body }),
    headers: {
      [HEADER_NAMES.accept]: JSON_MEDIA_TYPE,
      ...(spec.body === undefined ? {} : { [HEADER_NAMES.contentType]: JSON_MEDIA_TYPE }),
    },
    signal,
  });
  // The timer keeps running while the body downloads, so a stalled body also times out.
  const text = await response.text();
  return { response, text };
}

function classifyFetchFailure(
  cause: unknown,
  scope: AbortScope,
  callerSignal: AbortSignal | undefined,
  timeoutMs: number,
): Error {
  if (scope.timedOut()) {
    return new TimeoutError(timeoutMs, { cause });
  }
  if (callerSignal?.aborted) {
    return new CanceledError({ cause });
  }
  return new NetworkError({ cause });
}

function readBody<T>(
  { response, text }: { response: Response; text: string },
  isExpected: RequestSpec<T>['isExpected'],
): T {
  if (!response.ok) {
    throw buildApiError(response, text);
  }
  const body = parseJson(text);
  if (!isExpected(body)) {
    throw buildInvalidResponseError(response);
  }
  return body;
}

/** Creates a client. Tests pass their own config; the app uses the default client below. */
export function createApiClient(config: ApiClientConfig = {}): ApiClient {
  const baseUrl = resolveBaseUrl(config.baseUrl);
  const defaultTimeoutMs = config.timeoutMs ?? DEFAULT_REQUEST_TIMEOUT_MS;
  assertValidTimeout(defaultTimeoutMs);

  async function send<T>(spec: RequestSpec<T>, options: RequestOptions): Promise<T> {
    const timeoutMs = options.timeoutMs ?? defaultTimeoutMs;
    assertValidTimeout(timeoutMs);
    const callerSignal = options.signal;
    if (callerSignal?.aborted) {
      throw new CanceledError({ cause: callerSignal.reason });
    }

    const scope = createAbortScope(callerSignal, timeoutMs);
    let received: { response: Response; text: string };
    try {
      received = await fetchText(`${baseUrl}${spec.path}`, spec, scope.signal);
    } catch (cause) {
      throw classifyFetchFailure(cause, scope, callerSignal, timeoutMs);
    } finally {
      scope.dispose();
    }
    return readBody(received, spec.isExpected);
  }

  return {
    calculate: (request, options = {}) =>
      send(
        {
          path: API_PATHS.calculate,
          method: HTTP_METHODS.post,
          body: JSON.stringify(request),
          isExpected: isCalculateResponse,
        },
        options,
      ),
    getOperations: (options = {}) =>
      send(
        { path: API_PATHS.operations, method: HTTP_METHODS.get, isExpected: isOperationsResponse },
        options,
      ),
  };
}

export const defaultApiClient = createApiClient({ baseUrl: import.meta.env.VITE_API_BASE_URL });
