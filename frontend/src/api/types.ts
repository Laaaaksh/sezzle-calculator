// Mirrors docs/API.md. When the contract changes, change this file and the test that
// compares it with the document (types.test.ts) in the same commit.

/** Every operation name returned by GET /api/v1/operations. */
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

/** Every error code the server can send (the "Error codes and HTTP statuses" table). */
export const ERROR_CODES = [
  'BAD_REQUEST',
  'VALIDATION_FAILED',
  'UNSUPPORTED_MEDIA_TYPE',
  'PAYLOAD_TOO_LARGE',
  'NOT_FOUND',
  'METHOD_NOT_ALLOWED',
  'INTERNAL',
  'SERVICE_UNAVAILABLE',
  'UNKNOWN_OPERATION',
  'INVALID_OPERAND_COUNT',
  'OPERAND_TOO_LARGE',
  'INVALID_OPERAND',
  'DIVISION_BY_ZERO',
  'INVALID_EXPONENT',
  'EXPONENT_TOO_LARGE',
  'RESULT_TOO_LARGE',
  'NEGATIVE_SQUARE_ROOT',
] as const;

export type ErrorCode = (typeof ERROR_CODES)[number];

/**
 * Codes the client makes up itself. The server never sends them.
 * - INVALID_RESPONSE: the server answered, but the body does not match docs/API.md.
 * - UNKNOWN: an error response that is not in the standard error shape (a proxy's HTML page,
 *   say) or that carries a code this client does not know.
 */
export type ClientErrorCode = 'INVALID_RESPONSE' | 'UNKNOWN';

export type ApiErrorCode = ErrorCode | ClientErrorCode;

export interface CalculateRequest {
  operation: OperationType;
  /** Decimal strings, never JSON numbers: a float cannot hold what the user typed. */
  operands: string[];
}

export interface CalculateResponse {
  /** Repeats the request exactly as sent. */
  operation: OperationType;
  operands: string[];
  result: string;
}

export interface OperationInfo {
  name: OperationType;
  arity: number;
  /** For display only. Clients send the name, never the symbol. */
  symbol: string;
  description: string;
}

export interface OperationsResponse {
  operations: OperationInfo[];
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
