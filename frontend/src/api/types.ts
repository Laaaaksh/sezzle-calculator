// Mirrors docs/API.md. When the contract changes, change this file and the test that
// compares it with the document (types.test.ts) in the same commit.

/**
 * The operation names docs/API.md lists today. The server owns the real list, so request,
 * response and OperationInfo fields take any name and a new server operation needs no client change.
 */
export const OPERATION_TYPES = [
  'add',
  'subtract',
  'multiply',
  'divide',
  'power',
  'sqrt',
  'percentage',
] as const;

export type OperationType = (typeof OPERATION_TYPES)[number];

/** Every error code the server can send (the "Error codes and HTTP statuses" table), by name. */
export const ERROR_CODE = {
  BAD_REQUEST: 'BAD_REQUEST',
  VALIDATION_FAILED: 'VALIDATION_FAILED',
  UNSUPPORTED_MEDIA_TYPE: 'UNSUPPORTED_MEDIA_TYPE',
  PAYLOAD_TOO_LARGE: 'PAYLOAD_TOO_LARGE',
  NOT_FOUND: 'NOT_FOUND',
  METHOD_NOT_ALLOWED: 'METHOD_NOT_ALLOWED',
  INTERNAL: 'INTERNAL',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
  REQUEST_TIMEOUT: 'REQUEST_TIMEOUT',
  RATE_LIMITED: 'RATE_LIMITED',
  UNKNOWN_OPERATION: 'UNKNOWN_OPERATION',
  INVALID_OPERAND_COUNT: 'INVALID_OPERAND_COUNT',
  OPERAND_TOO_LARGE: 'OPERAND_TOO_LARGE',
  INVALID_OPERAND: 'INVALID_OPERAND',
  DIVISION_BY_ZERO: 'DIVISION_BY_ZERO',
  INVALID_EXPONENT: 'INVALID_EXPONENT',
  EXPONENT_TOO_LARGE: 'EXPONENT_TOO_LARGE',
  RESULT_TOO_LARGE: 'RESULT_TOO_LARGE',
  NEGATIVE_SQUARE_ROOT: 'NEGATIVE_SQUARE_ROOT',
} as const;

/** The same codes as a list, for runtime checks. */
export const ERROR_CODES = Object.values(ERROR_CODE);

export type ErrorCode = (typeof ERROR_CODE)[keyof typeof ERROR_CODE];

/**
 * Codes the client makes up itself. The server never sends them.
 * - INVALID_RESPONSE: the server answered, but the body does not match docs/API.md.
 * - UNKNOWN: an error response that is not in the standard error shape (a proxy's HTML page,
 *   say) or that carries a code this client does not know.
 */
export const CLIENT_ERROR_CODE = {
  INVALID_RESPONSE: 'INVALID_RESPONSE',
  UNKNOWN: 'UNKNOWN',
} as const;

export type ClientErrorCode = (typeof CLIENT_ERROR_CODE)[keyof typeof CLIENT_ERROR_CODE];

export type ApiErrorCode = ErrorCode | ClientErrorCode;

export interface CalculateRequest {
  /** One of the names from GET /api/v1/operations. */
  operation: string;
  /** Decimal strings, never JSON numbers: a float cannot hold what the user typed. */
  operands: string[];
}

export interface CalculateResponse {
  /** Repeats the request exactly as sent. */
  operation: string;
  operands: string[];
  result: string;
}

export interface OperationInfo {
  /** The value to send as CalculateRequest.operation. */
  name: string;
  arity: number;
  /** For display only. Clients send the name, never the symbol. */
  symbol: string;
  description: string;
}

/**
 * The server's input limits, sent with the operations list. The UI uses maxOperandLength to stop
 * typing at the same length the server accepts. The other two are carried for completeness.
 */
export interface OperationLimits {
  maxOperandLength: number;
  maxExponent: number;
  decimalPrecision: number;
}

export interface OperationsResponse {
  operations: OperationInfo[];
  /** Absent from an older server: callers then fall back to their built-in defaults. */
  limits?: OperationLimits;
}

/** Values in `fields` are strings or numbers. The keys are fixed per code. */
export type ErrorFields = Record<string, string | number>;

export interface ApiErrorBody {
  error: {
    code: ErrorCode;
    message: string;
    requestId: string;
    fields?: ErrorFields;
  };
}
