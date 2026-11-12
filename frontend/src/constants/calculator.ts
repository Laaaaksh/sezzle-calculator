import { OPERATION_TYPES, type OperationType } from '../api';

/**
 * Longest operand the keypad accepts, counting "-" and ".". Matches the backend default for
 * `max_operand_length` in docs/API.md (64). If the server is configured with a lower limit the
 * request is rejected with OPERAND_TOO_LARGE and the UI shows that error.
 */
export const MAX_OPERAND_LENGTH = 64;

/** How many entries the history list keeps. The oldest one is dropped first. */
export const HISTORY_LIMIT = 20;

export const DIGITS = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'] as const;
export type Digit = (typeof DIGITS)[number];

/** Operations that take one operand. Everything else in OperationType takes two. */
export const UNARY_OPERATIONS = ['sqrt'] as const;
export type UnaryOperation = (typeof UNARY_OPERATIONS)[number];
export type BinaryOperation = Exclude<OperationType, UnaryOperation>;

export function isBinaryOperation(operation: OperationType): operation is BinaryOperation {
  return !UNARY_OPERATIONS.some((unary) => unary === operation);
}

export const BINARY_OPERATIONS: readonly BinaryOperation[] =
  OPERATION_TYPES.filter(isBinaryOperation);
