import { MAX_OPERAND_LENGTH, type BinaryOperation } from '../../../constants/calculator';
import {
  CLIENT_FAILURE_CODE,
  type CalculatorAction,
  type CalculatorError,
  type CalculatorState,
  type EnteringFirstState,
  type EnteringSecondState,
  type PendingRequest,
  type ShowingResultState,
  type WaitingState,
} from './types';

const EMPTY_OPERAND = '';
const ZERO = '0';
const DECIMAL_POINT = '.';
const MINUS = '-';

/**
 * A decimal point needs itself plus one digit ("0." when empty), so it is refused when fewer than
 * this many characters remain.
 */
const DECIMAL_POINT_RESERVE = 2;

/** Any non-zero digit. Not global, so .test() keeps no state between calls. */
const NON_ZERO_DIGIT = /[1-9]/;

export interface ReducerLimits {
  maxOperandLength: number;
}

const DEFAULT_LIMITS: ReducerLimits = { maxOperandLength: MAX_OPERAND_LENGTH };

export function createInitialState(lastRequestId = 0): CalculatorState {
  return { status: 'entering_first', operand: EMPTY_OPERAND, lastRequestId };
}

export const initialState: CalculatorState = createInitialState();

// Operands are strings from start to end. A float cannot hold what the user typed
// (see "Numbers are strings" in docs/API.md), so nothing here calls Number() or parseFloat().

function appendChar(operand: string, char: string, max: number): string {
  if (char === DECIMAL_POINT) {
    if (operand.includes(DECIMAL_POINT) || operand.length + DECIMAL_POINT_RESERVE > max) {
      return operand;
    }
    // A bare "." is not a valid operand, so an empty operand becomes "0.".
    return operand === EMPTY_OPERAND ? `${ZERO}${DECIMAL_POINT}` : `${operand}${DECIMAL_POINT}`;
  }
  if (operand.length >= max) {
    return operand;
  }
  // Leading zeros collapse: typing after a lone 0 (or -0) replaces it, and 0 then 0 stays 0.
  if (operand === ZERO) {
    return char;
  }
  if (operand === `${MINUS}${ZERO}`) {
    return `${MINUS}${char}`;
  }
  return `${operand}${char}`;
}

function isZero(operand: string): boolean {
  return !NON_ZERO_DIGIT.test(operand);
}

/**
 * Flips the sign. Zero stays unsigned; the added "-" counts toward the limit, so an operand at the
 * limit comes back unchanged, which callers treat as a refusal.
 */
function negate(operand: string, max: number): string {
  if (isZero(operand)) {
    return operand;
  }
  if (operand.startsWith(MINUS)) {
    return operand.slice(1);
  }
  return operand.length >= max ? operand : `${MINUS}${operand}`;
}

function finalizeOperand(operand: string): string {
  const trimmed = operand.endsWith(DECIMAL_POINT) ? operand.slice(0, -1) : operand;
  return trimmed === EMPTY_OPERAND || trimmed === `${MINUS}${ZERO}` ? ZERO : trimmed;
}

function removeLastChar(operand: string): string {
  const shorter = operand.slice(0, -1);
  return shorter === MINUS ? EMPTY_OPERAND : shorter;
}

function tooLongToReuse(result: string, max: number): boolean {
  return result.length > max;
}

/**
 * Keeps the result on screen (it can still be read and copied) and says why it cannot feed
 * another calculation. Without this guard the server would reject the next request.
 */
function refuseLongResult(state: ShowingResultState): ShowingResultState {
  return {
    ...state,
    notice: {
      code: CLIENT_FAILURE_CODE.RESULT_TOO_LONG_TO_REUSE,
      requestId: undefined,
      retryable: false,
    } satisfies CalculatorError,
  };
}

function startRequest(
  state: CalculatorState,
  request: Pick<PendingRequest, 'operation' | 'operands' | 'then'>,
): WaitingState {
  const id = state.lastRequestId + 1;
  return { status: 'waiting', lastRequestId: id, pending: { id, ...request } };
}

function typeChar(state: CalculatorState, char: string, max: number): CalculatorState {
  switch (state.status) {
    case 'entering_first':
    case 'entering_second': {
      const operand = appendChar(state.operand, char, max);
      return operand === state.operand ? state : { ...state, operand };
    }
    case 'operator_chosen':
      return {
        status: 'entering_second',
        lastRequestId: state.lastRequestId,
        left: state.left,
        operator: state.operator,
        operand: appendChar(EMPTY_OPERAND, char, max),
      };
    case 'showing_result':
    case 'error':
      return {
        status: 'entering_first',
        lastRequestId: state.lastRequestId,
        operand: appendChar(EMPTY_OPERAND, char, max),
      };
    case 'waiting':
      return state;
  }
}

function chooseOperator(
  state: CalculatorState,
  operator: BinaryOperation,
  max: number,
): CalculatorState {
  switch (state.status) {
    case 'entering_first':
      return {
        status: 'operator_chosen',
        lastRequestId: state.lastRequestId,
        left: finalizeOperand(state.operand),
        operator,
      };
    case 'operator_chosen':
      return state.operator === operator ? state : { ...state, operator };
    case 'entering_second':
      return startRequest(state, {
        operation: state.operator,
        operands: [state.left, finalizeOperand(state.operand)],
        then: operator,
      });
    case 'showing_result':
      return tooLongToReuse(state.result, max)
        ? refuseLongResult(state)
        : {
            status: 'operator_chosen',
            lastRequestId: state.lastRequestId,
            left: state.result,
            operator,
          };
    case 'waiting':
    case 'error':
      return state;
  }
}

function applyUnary(state: CalculatorState, operation: 'sqrt', max: number): CalculatorState {
  switch (state.status) {
    case 'entering_first':
      return state.operand === EMPTY_OPERAND
        ? state
        : startRequest(state, {
            operation,
            operands: [finalizeOperand(state.operand)],
            then: null,
          });
    case 'showing_result':
      return tooLongToReuse(state.result, max)
        ? refuseLongResult(state)
        : startRequest(state, { operation, operands: [state.result], then: null });
    default:
      // With an operator pending, sqrt is ambiguous (of which operand?), so it is ignored.
      return state;
  }
}

function equals(state: CalculatorState): CalculatorState {
  if (state.status !== 'entering_second') {
    return state;
  }
  return startRequest(state, {
    operation: state.operator,
    operands: [state.left, finalizeOperand(state.operand)],
    then: null,
  });
}

function backspace(state: CalculatorState): CalculatorState {
  switch (state.status) {
    case 'entering_first':
      return state.operand === EMPTY_OPERAND
        ? state
        : { ...state, operand: removeLastChar(state.operand) };
    case 'operator_chosen':
      return {
        status: 'entering_first',
        lastRequestId: state.lastRequestId,
        operand: state.left,
      };
    case 'entering_second': {
      const operand = removeLastChar(state.operand);
      return operand === EMPTY_OPERAND
        ? {
            status: 'operator_chosen',
            lastRequestId: state.lastRequestId,
            left: state.left,
            operator: state.operator,
          }
        : { ...state, operand };
    }
    case 'showing_result':
    case 'error':
      // A result or an error is not text to edit, so backspace dismisses it like clear.
      return createInitialState(state.lastRequestId);
    case 'waiting':
      return state;
  }
}

function toggleSign(state: CalculatorState, max: number): CalculatorState {
  switch (state.status) {
    case 'entering_first':
    case 'entering_second': {
      const operand = negate(state.operand, max);
      return operand === state.operand ? state : { ...state, operand };
    }
    case 'showing_result': {
      // The negated result becomes an operand, so the same length rule applies.
      if (tooLongToReuse(state.result, max)) {
        return refuseLongResult(state);
      }
      const operand = negate(state.result, max);
      if (operand === state.result) {
        return isZero(state.result) ? state : refuseLongResult(state);
      }
      return {
        status: 'entering_first',
        lastRequestId: state.lastRequestId,
        operand,
      } satisfies EnteringFirstState;
    }
    default:
      return state;
  }
}

function resultReceived(
  state: CalculatorState,
  requestId: number,
  result: string,
  max: number,
): CalculatorState {
  // Only the answer to the request we are waiting for counts. Anything else is stale.
  if (state.status !== 'waiting' || state.pending.id !== requestId) {
    return state;
  }
  const { operation, operands, then } = state.pending;
  const shown: ShowingResultState = {
    status: 'showing_result',
    lastRequestId: state.lastRequestId,
    result,
    request: { operation, operands },
  };
  if (then !== null && tooLongToReuse(result, max)) {
    // The chained operator would send this result as an operand. Show it instead, with the reason.
    return refuseLongResult(shown);
  }
  if (then !== null) {
    return {
      status: 'operator_chosen',
      lastRequestId: state.lastRequestId,
      left: result,
      operator: then,
    };
  }
  return shown;
}

function reuse(state: CalculatorState, result: string, max: number): CalculatorState {
  if (state.status === 'waiting') {
    // A request is in flight; the answer would overwrite whatever was reused.
    return state;
  }
  if (tooLongToReuse(result, max)) {
    // Not usable as an operand, so it is shown for reading only. Any calculation in progress
    // is dropped, which the message makes clear by replacing the expression.
    return refuseLongResult({
      status: 'showing_result',
      lastRequestId: state.lastRequestId,
      result,
      request: null,
    });
  }
  switch (state.status) {
    case 'operator_chosen':
    case 'entering_second':
      return {
        status: 'entering_second',
        lastRequestId: state.lastRequestId,
        left: state.left,
        operator: state.operator,
        operand: result,
      } satisfies EnteringSecondState;
    default:
      return {
        status: 'showing_result',
        lastRequestId: state.lastRequestId,
        result,
        request: null,
      };
  }
}

/**
 * Limits are fixed per reducer, so the hook builds a new one when the server reports different
 * values.
 */
export function createCalculatorReducer({ maxOperandLength }: ReducerLimits) {
  return (state: CalculatorState, action: CalculatorAction): CalculatorState =>
    reduce(state, action, maxOperandLength);
}

export const calculatorReducer = createCalculatorReducer(DEFAULT_LIMITS);

function reduce(state: CalculatorState, action: CalculatorAction, max: number): CalculatorState {
  switch (action.type) {
    case 'digit':
      return typeChar(state, action.digit, max);
    case 'decimal':
      return typeChar(state, DECIMAL_POINT, max);
    case 'operator':
      return chooseOperator(state, action.operator, max);
    case 'unary':
      return applyUnary(state, action.operation, max);
    case 'equals':
      return equals(state);
    case 'backspace':
      return backspace(state);
    case 'clear':
      // lastRequestId is kept so a request sent before the clear can never match a later one.
      return createInitialState(state.lastRequestId);
    case 'toggle_sign':
      return toggleSign(state, max);
    case 'result_received':
      return resultReceived(state, action.requestId, action.result, max);
    case 'request_failed':
      return state.status === 'waiting' && state.pending.id === action.requestId
        ? {
            status: 'error',
            lastRequestId: state.lastRequestId,
            error: action.error,
            pending: state.pending,
          }
        : state;
    case 'retry':
      return state.status === 'error' && state.error.retryable
        ? startRequest(state, {
            operation: state.pending.operation,
            operands: state.pending.operands,
            then: state.pending.then,
          })
        : state;
    case 'reuse':
      return reuse(state, action.result, max);
  }
}
