import type { OperationType } from '../api';

/**
 * Longest operand the keypad accepts, counting "-" and ".", when the server does not say. The
 * server reports its real limit as limits.maxOperandLength in GET /api/v1/operations and that
 * value wins; this is the fallback for an older server or a failed lookup. It matches the backend
 * default for `max_operand_length` in docs/API.md (64).
 */
export const MAX_OPERAND_LENGTH = 64;

/** How many entries the history list keeps. The oldest one is dropped first. */
export const HISTORY_LIMIT = 20;

export const DIGITS = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'] as const;
export type Digit = (typeof DIGITS)[number];

/** Operations that take one operand. Everything else in OperationType takes two. */
export type UnaryOperation = Extract<OperationType, 'sqrt'>;
export type BinaryOperation = Exclude<OperationType, UnaryOperation>;

/** Symbols for showing an expression. Display only: requests carry the operation name. */
export const OPERATION_SYMBOLS: Record<OperationType, string> = {
  add: '+',
  subtract: '-',
  multiply: '×',
  divide: '÷',
  power: '^',
  sqrt: '√',
  percentage: '%',
};
