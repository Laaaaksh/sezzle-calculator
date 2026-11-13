import type { ApiErrorCode, OperationType } from '../../../api';
import type { BinaryOperation, Digit, UnaryOperation } from '../../../constants/calculator';

export const CLIENT_FAILURE_CODE = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT: 'TIMEOUT',
  RESULT_TOO_LONG_TO_REUSE: 'RESULT_TOO_LONG_TO_REUSE',
} as const;

export type ClientFailureCode = (typeof CLIENT_FAILURE_CODE)[keyof typeof CLIENT_FAILURE_CODE];

export interface CalculatorError {
  /** The text to show comes from constants/errorMessages.ts, keyed by this code. */
  code: ApiErrorCode | ClientFailureCode;
  /** Quote this to support. Absent when no response arrived. */
  requestId: string | undefined;
  retryable: boolean;
}

export interface CalculationRequest {
  operation: OperationType;
  operands: readonly string[];
}

export interface PendingRequest extends CalculationRequest {
  /** Responses carry this id so a late answer to an older request can be ignored. */
  id: number;
  /** An operator pressed while a second operand was shown: it continues from the result. */
  then: BinaryOperation | null;
}

interface StateBase {
  /** The id of the newest request ever issued. Only ever grows, even across clear. */
  lastRequestId: number;
}

export interface EnteringFirstState extends StateBase {
  status: 'entering_first';
  operand: string;
}

export interface OperatorChosenState extends StateBase {
  status: 'operator_chosen';
  left: string;
  operator: BinaryOperation;
}

export interface EnteringSecondState extends StateBase {
  status: 'entering_second';
  left: string;
  operator: BinaryOperation;
  /** Never empty: backspacing the last character returns to operator_chosen. */
  operand: string;
}

export interface WaitingState extends StateBase {
  status: 'waiting';
  pending: PendingRequest;
}

export interface ShowingResultState extends StateBase {
  status: 'showing_result';
  result: string;
  /** Null when the result was reused from history, so there is no expression to show. */
  request: CalculationRequest | null;
  /** Why the last key was refused (result too long to reuse). The next transition clears it. */
  notice?: CalculatorError | undefined;
}

export interface ErrorState extends StateBase {
  status: 'error';
  error: CalculatorError;
  pending: PendingRequest;
}

export type CalculatorState =
  | EnteringFirstState
  | OperatorChosenState
  | EnteringSecondState
  | WaitingState
  | ShowingResultState
  | ErrorState;

export type CalculatorAction =
  | { type: 'digit'; digit: Digit }
  | { type: 'decimal' }
  | { type: 'operator'; operator: BinaryOperation }
  | { type: 'unary'; operation: UnaryOperation }
  | { type: 'equals' }
  | { type: 'backspace' }
  | { type: 'clear' }
  | { type: 'toggle_sign' }
  | { type: 'result_received'; requestId: number; result: string }
  | { type: 'request_failed'; requestId: number; error: CalculatorError }
  | { type: 'retry' }
  | { type: 'reuse'; result: string };
