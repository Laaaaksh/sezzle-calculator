import type { OperationType } from '../api';

/**
 * Longest operand the keypad accepts ("-" and "." count) when the server reports no limit. The
 * server's limits.maxOperandLength from GET /api/v1/operations wins; this fallback matches the
 * `max_operand_length` default (64) in docs/API.md.
 */
export const MAX_OPERAND_LENGTH = 64;

export const HISTORY_LIMIT = 20;

export const DIGITS = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'] as const;
export type Digit = (typeof DIGITS)[number];

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
