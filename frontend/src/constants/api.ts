/** Paths of the backend routes, as listed in docs/API.md. */
export const API_PATHS = {
  calculate: '/api/v1/calculate',
  operations: '/api/v1/operations',
} as const;

/** How long one request may take, in milliseconds, before it fails with a TimeoutError. */
export const DEFAULT_REQUEST_TIMEOUT_MS = 8000;

/** Used when VITE_API_BASE_URL is unset or empty: same origin, so the dev proxy applies. */
export const DEFAULT_API_BASE_URL = '';

export const HTTP_METHODS = {
  get: 'GET',
  post: 'POST',
} as const;

export const HEADER_NAMES = {
  accept: 'Accept',
  contentType: 'Content-Type',
  requestId: 'X-Request-ID',
} as const;

export const JSON_MEDIA_TYPE = 'application/json';
