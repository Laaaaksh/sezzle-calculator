import {
  ERROR_CODES,
  OPERATION_TYPES,
  type ApiErrorBody,
  type CalculateResponse,
  type ErrorCode,
  type ErrorFields,
  type OperationInfo,
  type OperationType,
  type OperationsResponse,
} from './types';

// Small runtime checks for data that crossed the network. TypeScript types vanish at runtime,
// so without these a server change would surface as a confusing crash far from the cause.

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === 'string');
}

export function isOperationType(value: unknown): value is OperationType {
  return OPERATION_TYPES.some((operation) => operation === value);
}

export function isErrorCode(value: unknown): value is ErrorCode {
  return ERROR_CODES.some((code) => code === value);
}

function isErrorFields(value: unknown): value is ErrorFields {
  return (
    isRecord(value) &&
    Object.values(value).every((item) => typeof item === 'string' || typeof item === 'number')
  );
}

export function isCalculateResponse(value: unknown): value is CalculateResponse {
  return (
    isRecord(value) &&
    isOperationType(value.operation) &&
    isStringArray(value.operands) &&
    typeof value.result === 'string'
  );
}

function isOperationInfo(value: unknown): value is OperationInfo {
  return (
    isRecord(value) &&
    isOperationType(value.name) &&
    typeof value.arity === 'number' &&
    Number.isInteger(value.arity) &&
    value.arity > 0 &&
    typeof value.symbol === 'string' &&
    typeof value.description === 'string'
  );
}

export function isOperationsResponse(value: unknown): value is OperationsResponse {
  return (
    isRecord(value) && Array.isArray(value.operations) && value.operations.every(isOperationInfo)
  );
}

/**
 * Checks the standard error shape. The code is checked only as a string here, so an error
 * with a code from a newer server still gives the caller its message and request ID.
 */
export function isErrorBodyShape(
  value: unknown,
): value is { error: Omit<ApiErrorBody['error'], 'code'> & { code: string } } {
  if (!isRecord(value) || !isRecord(value.error)) {
    return false;
  }
  const { code, message, requestId, fields } = value.error;
  return (
    typeof code === 'string' &&
    typeof message === 'string' &&
    typeof requestId === 'string' &&
    (fields === undefined || isErrorFields(fields))
  );
}
