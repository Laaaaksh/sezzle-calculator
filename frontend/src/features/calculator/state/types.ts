import type { ApiErrorCode, OperationType } from '../../../api';
import type { BinaryOperation, Digit, UnaryOperation } from '../../../constants/calculator';

/** Codes the UI makes up for failures that have no server response. */
export type ClientFailureCode = 'NETWORK_ERROR' | 'TIMEOUT';

export interface CalculatorError {
  code: ApiErrorCode | ClientFailureCode;
  message: string;
  /** Quote this to support. Absent when no response arrived. */
  requestId: string | undefined;
  /** True when sending the same request again can succeed (network and timeout failures). */
  retryable: boolean;
}

/** What was calculated: the operation and its operands, exactly as sent to the server. */
export interface CalculationRequest {
  operation: OperationType;
  operands: readonly string[];
}

/** A request the reducer wants sent. The hook sends it and reports back with the same id. */
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
  /** Digits typed so far. Empty means nothing typed yet (shown as 0). */
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
  /** The server's string, never parsed into a JS number. */
  result: string;
  /** Null when the result was reused from history, so there is no expression to show. */
  request: CalculationRequest | null;
}

export interface ErrorState extends StateBase {
  status: 'error';
  error: CalculatorError;
  /** The request that failed, kept so a retry can send it again. */
  pending: PendingRequest;
}

export type CalculatorState =
  | EnteringFirstState
  | OperatorChosenState
  | EnteringSecondState
  | WaitingState
  | ShowingResultState
  | ErrorState;

export type CalculatorStatus = CalculatorState['status'];

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
