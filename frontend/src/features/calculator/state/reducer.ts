import { MAX_OPERAND_LENGTH, type BinaryOperation } from '../../../constants/calculator';
import type {
  CalculatorAction,
  CalculatorState,
  EnteringFirstState,
  EnteringSecondState,
  PendingRequest,
  WaitingState,
} from './types';

/** Operand text when nothing has been typed. Shown as 0 and sent as "0". */
const EMPTY_OPERAND = '';
const ZERO = '0';
const DECIMAL_POINT = '.';
const MINUS = '-';

export function createInitialState(lastRequestId = 0): CalculatorState {
  return { status: 'entering_first', operand: EMPTY_OPERAND, lastRequestId };
}

export const initialState: CalculatorState = createInitialState();

// Operands are strings from start to end. A float cannot hold what the user typed
// (see "Numbers are strings" in docs/API.md), so nothing here calls Number() or parseFloat().

/** Adds one character to an operand, or returns the operand unchanged when the rule forbids it. */
function appendChar(operand: string, char: string): string {
  if (char === DECIMAL_POINT) {
    if (operand.includes(DECIMAL_POINT) || operand.length + 2 > MAX_OPERAND_LENGTH) {
      return operand;
    }
    // A bare "." is not a valid operand, so an empty operand becomes "0.".
    return operand === EMPTY_OPERAND ? `${ZERO}${DECIMAL_POINT}` : `${operand}${DECIMAL_POINT}`;
  }
  if (operand.length >= MAX_OPERAND_LENGTH) {
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

/** Flips the sign. A zero stays unsigned, and the limit still counts the added "-". */
function negate(operand: string): string {
  if (!/[1-9]/.test(operand)) {
    return operand;
  }
  if (operand.startsWith(MINUS)) {
    return operand.slice(1);
  }
  return operand.length >= MAX_OPERAND_LENGTH ? operand : `${MINUS}${operand}`;
}

/** The operand as it is sent: no dangling ".", and an empty or "-0" operand is 0. */
function finalizeOperand(operand: string): string {
  const trimmed = operand.endsWith(DECIMAL_POINT) ? operand.slice(0, -1) : operand;
  return trimmed === EMPTY_OPERAND || trimmed === `${MINUS}${ZERO}` ? ZERO : trimmed;
}

function removeLastChar(operand: string): string {
  const shorter = operand.slice(0, -1);
  return shorter === MINUS ? EMPTY_OPERAND : shorter;
}

function startRequest(
  state: CalculatorState,
  request: Pick<PendingRequest, 'operation' | 'operands' | 'then'>,
): WaitingState {
  const id = state.lastRequestId + 1;
  return { status: 'waiting', lastRequestId: id, pending: { id, ...request } };
}

function typeChar(state: CalculatorState, char: string): CalculatorState {
  switch (state.status) {
    case 'entering_first':
    case 'entering_second': {
      const operand = appendChar(state.operand, char);
      return operand === state.operand ? state : { ...state, operand };
    }
    case 'operator_chosen':
      return {
        status: 'entering_second',
        lastRequestId: state.lastRequestId,
        left: state.left,
        operator: state.operator,
        operand: appendChar(EMPTY_OPERAND, char),
      };
    case 'showing_result':
    case 'error':
      // Typing after an answer starts a new calculation.
      return {
        status: 'entering_first',
        lastRequestId: state.lastRequestId,
        operand: appendChar(EMPTY_OPERAND, char),
      };
    case 'waiting':
      return state;
  }
}

function chooseOperator(state: CalculatorState, operator: BinaryOperation): CalculatorState {
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
      // Like a desk calculator: compute what is entered, then continue from its result.
      return startRequest(state, {
        operation: state.operator,
        operands: [state.left, finalizeOperand(state.operand)],
        then: operator,
      });
    case 'showing_result':
      return {
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

function applyUnary(state: CalculatorState, operation: 'sqrt'): CalculatorState {
  switch (state.status) {
    case 'entering_first':
      // Nothing typed means nothing to compute.
      return state.operand === EMPTY_OPERAND
        ? state
        : startRequest(state, {
            operation,
            operands: [finalizeOperand(state.operand)],
            then: null,
          });
    case 'showing_result':
      return startRequest(state, { operation, operands: [state.result], then: null });
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
      // Removing the operator goes back to editing the first operand.
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

function toggleSign(state: CalculatorState): CalculatorState {
  switch (state.status) {
    case 'entering_first':
    case 'entering_second': {
      const operand = negate(state.operand);
      return operand === state.operand ? state : { ...state, operand };
    }
    case 'showing_result': {
      const operand = negate(state.result);
      return operand === state.result
        ? state
        : ({
            status: 'entering_first',
            lastRequestId: state.lastRequestId,
            operand,
          } satisfies EnteringFirstState);
    }
    default:
      return state;
  }
}

function resultReceived(
  state: CalculatorState,
  requestId: number,
  result: string,
): CalculatorState {
  // Only the answer to the request we are waiting for counts. Anything else is stale.
  if (state.status !== 'waiting' || state.pending.id !== requestId) {
    return state;
  }
  const { operation, operands, then } = state.pending;
  if (then !== null) {
    return {
      status: 'operator_chosen',
      lastRequestId: state.lastRequestId,
      left: result,
      operator: then,
    };
  }
  return {
    status: 'showing_result',
    lastRequestId: state.lastRequestId,
    result,
    request: { operation, operands },
  };
}

function reuse(state: CalculatorState, result: string): CalculatorState {
  switch (state.status) {
    case 'waiting':
      // A request is in flight; the answer would overwrite whatever was reused.
      return state;
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

export function calculatorReducer(
  state: CalculatorState,
  action: CalculatorAction,
): CalculatorState {
  switch (action.type) {
    case 'digit':
      return typeChar(state, action.digit);
    case 'decimal':
      return typeChar(state, DECIMAL_POINT);
    case 'operator':
      return chooseOperator(state, action.operator);
    case 'unary':
      return applyUnary(state, action.operation);
    case 'equals':
      return equals(state);
    case 'backspace':
      return backspace(state);
    case 'clear':
      // lastRequestId is kept so a request sent before the clear can never match a later one.
      return createInitialState(state.lastRequestId);
    case 'toggle_sign':
      return toggleSign(state);
    case 'result_received':
      return resultReceived(state, action.requestId, action.result);
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
      return reuse(state, action.result);
  }
}
