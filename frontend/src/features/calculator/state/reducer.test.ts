import { describe, expect, it } from 'vitest';
import { DIGITS, MAX_OPERAND_LENGTH, type BinaryOperation } from '../../../constants/calculator';
import {
  calculatorReducer,
  createCalculatorReducer,
  createInitialState,
  initialState,
} from './reducer';
import type { CalculatorAction, CalculatorError, CalculatorState } from './types';

const NETWORK_ERROR: CalculatorError = {
  code: 'NETWORK_ERROR',
  requestId: undefined,
  retryable: true,
};

const API_ERROR: CalculatorError = {
  code: 'DIVISION_BY_ZERO',
  requestId: 'req-1',
  retryable: false,
};

function run(state: CalculatorState, ...actions: CalculatorAction[]): CalculatorState {
  return actions.reduce(calculatorReducer, state);
}

function keys(text: string): CalculatorAction[] {
  const actions: CalculatorAction[] = [];
  for (const char of text) {
    const digit = DIGITS.find((candidate) => candidate === char);
    if (char === '.') {
      actions.push({ type: 'decimal' });
    } else if (digit !== undefined) {
      actions.push({ type: 'digit', digit });
    } else {
      throw new Error(`keys() cannot press "${char}"`);
    }
  }
  return actions;
}

function typed(text: string): CalculatorState {
  return run(initialState, ...keys(text));
}

function op(operator: BinaryOperation): CalculatorAction {
  return { type: 'operator', operator };
}

const EQUALS: CalculatorAction = { type: 'equals' };
const CLEAR: CalculatorAction = { type: 'clear' };
const BACKSPACE: CalculatorAction = { type: 'backspace' };
const TOGGLE: CalculatorAction = { type: 'toggle_sign' };
const SQRT: CalculatorAction = { type: 'unary', operation: 'sqrt' };

function waitingState(): CalculatorState {
  return run(typed('12'), op('add'), ...keys('7'), EQUALS);
}

function resultState(result = '19'): CalculatorState {
  return run(waitingState(), { type: 'result_received', requestId: 1, result });
}

function errorState(error: CalculatorError = API_ERROR): CalculatorState {
  return run(waitingState(), { type: 'request_failed', requestId: 1, error });
}

const SECOND_OPERAND_STATE = run(typed('12'), op('add'), ...keys('7'));

describe('initial state', () => {
  it('starts entering the first operand with nothing typed', () => {
    expect(initialState).toEqual({ status: 'entering_first', operand: '', lastRequestId: 0 });
  });
});

describe('digits', () => {
  it('appends digits to the first operand', () => {
    expect(typed('123')).toMatchObject({ status: 'entering_first', operand: '123' });
  });

  it('starts the second operand when a digit follows an operator', () => {
    expect(run(typed('1'), op('add'), ...keys('2'))).toMatchObject({
      status: 'entering_second',
      left: '1',
      operator: 'add',
      operand: '2',
    });
  });

  it('starts a new calculation when a digit follows a result', () => {
    expect(run(resultState(), ...keys('5'))).toMatchObject({
      status: 'entering_first',
      operand: '5',
    });
  });

  it('starts a new calculation when a digit follows an error', () => {
    expect(run(errorState(), ...keys('5'))).toMatchObject({
      status: 'entering_first',
      operand: '5',
    });
  });

  it('ignores digits while a request is pending', () => {
    const waiting = waitingState();

    expect(run(waiting, ...keys('5'))).toBe(waiting);
  });
});

describe('leading zeros', () => {
  it('collapses repeated zeros into one', () => {
    expect(typed('000')).toMatchObject({
      operand: '0',
    });
  });

  it('replaces a lone zero with the next digit', () => {
    expect(typed('05')).toMatchObject({ operand: '5' });
  });

  it('keeps zeros after the decimal point', () => {
    expect(typed('0.05')).toMatchObject({ operand: '0.05' });
  });

  it('keeps zeros inside a number', () => {
    expect(typed('100')).toMatchObject({ operand: '100' });
  });

  it('collapses leading zeros in the second operand', () => {
    expect(run(typed('1'), op('add'), ...keys('007'))).toMatchObject({ operand: '7' });
  });

  it('replaces a negative zero with the next digit', () => {
    expect(run(typed('0.5'), TOGGLE, BACKSPACE, BACKSPACE, ...keys('7'))).toMatchObject({
      operand: '-7',
    });
  });
});

describe('decimal point', () => {
  it('allows only one decimal point per operand', () => {
    expect(typed('1.2.3')).toMatchObject({ operand: '1.23' });
  });

  it('allows one decimal point in each operand', () => {
    expect(run(typed('1.5'), op('add'), ...keys('2.5'))).toMatchObject({
      left: '1.5',
      operand: '2.5',
    });
  });

  it('turns a decimal point on an empty operand into 0.', () => {
    expect(typed('.')).toMatchObject({ operand: '0.' });
  });

  it('turns a decimal point right after an operator into 0. in the second operand', () => {
    expect(run(typed('3'), op('add'), { type: 'decimal' })).toMatchObject({
      status: 'entering_second',
      operand: '0.',
    });
  });

  it('starts a new calculation with 0. when a decimal point follows a result', () => {
    expect(run(resultState(), { type: 'decimal' })).toMatchObject({
      status: 'entering_first',
      operand: '0.',
    });
  });

  it('returns the same state when the decimal point is refused', () => {
    const state = typed('1.5');

    expect(run(state, { type: 'decimal' })).toBe(state);
  });
});

describe('maximum operand length', () => {
  const full = '1'.repeat(MAX_OPERAND_LENGTH);

  it('accepts digits up to the limit', () => {
    expect(typed(full)).toMatchObject({ operand: full });
  });

  it('ignores a digit past the limit', () => {
    const state = typed(full);

    expect(run(state, ...keys('2'))).toBe(state);
  });

  it('ignores a digit past the limit in the second operand', () => {
    const state = run(typed('1'), op('add'), ...keys(full));

    expect(run(state, ...keys('2'))).toBe(state);
  });

  it('ignores a decimal point that would pass the limit', () => {
    const state = typed(full);

    expect(run(state, { type: 'decimal' })).toBe(state);
  });

  it('ignores a decimal point one character short of the limit, which leaves no room for a digit', () => {
    const state = typed('1'.repeat(MAX_OPERAND_LENGTH - 1));

    expect(run(state, { type: 'decimal' })).toBe(state);
  });

  it('accepts a decimal point with room for a digit after it', () => {
    const state = typed('1'.repeat(MAX_OPERAND_LENGTH - 2));

    expect(run(state, { type: 'decimal' })).toMatchObject({
      operand: `${'1'.repeat(MAX_OPERAND_LENGTH - 2)}.`,
    });
  });

  it('counts the minus sign toward the limit', () => {
    const state = run(typed('1'.repeat(MAX_OPERAND_LENGTH - 1)), TOGGLE);

    expect(state).toMatchObject({ operand: `-${'1'.repeat(MAX_OPERAND_LENGTH - 1)}` });
    expect(run(state, ...keys('2'))).toBe(state);
  });

  it('refuses a minus sign that would pass the limit', () => {
    const state = typed(full);

    expect(run(state, TOGGLE)).toBe(state);
  });
});

describe('operators', () => {
  it('moves to operator chosen with the first operand', () => {
    expect(run(typed('12'), op('multiply'))).toMatchObject({
      status: 'operator_chosen',
      left: '12',
      operator: 'multiply',
    });
  });

  it('uses 0 as the first operand when nothing was typed', () => {
    expect(run(initialState, op('subtract'))).toMatchObject({ left: '0', operator: 'subtract' });
  });

  it('drops a dangling decimal point from the first operand', () => {
    expect(run(typed('5.'), op('add'))).toMatchObject({ left: '5' });
  });

  it('treats a negative zero first operand as 0', () => {
    expect(run(typed('0.5'), TOGGLE, BACKSPACE, BACKSPACE, op('add'))).toMatchObject({ left: '0' });
  });

  it('replaces the operator when pressed twice', () => {
    expect(run(typed('12'), op('add'), op('divide'))).toMatchObject({
      status: 'operator_chosen',
      left: '12',
      operator: 'divide',
    });
  });

  it('returns the same state when the same operator is pressed again', () => {
    const state = run(typed('12'), op('add'));

    expect(run(state, op('add'))).toBe(state);
  });

  it('chains from the result when an operator follows a result', () => {
    expect(run(resultState('19.75'), op('multiply'))).toMatchObject({
      status: 'operator_chosen',
      left: '19.75',
      operator: 'multiply',
    });
  });

  it('computes the pending operation first when an operator follows a second operand', () => {
    expect(run(SECOND_OPERAND_STATE, op('multiply'))).toMatchObject({
      status: 'waiting',
      pending: { id: 1, operation: 'add', operands: ['12', '7'], then: 'multiply' },
    });
  });

  it('continues from the result with the queued operator', () => {
    const state = run(SECOND_OPERAND_STATE, op('multiply'), {
      type: 'result_received',
      requestId: 1,
      result: '19',
    });

    expect(state).toMatchObject({ status: 'operator_chosen', left: '19', operator: 'multiply' });
  });

  it('ignores an operator while a request is pending', () => {
    const waiting = waitingState();

    expect(run(waiting, op('add'))).toBe(waiting);
  });

  it('ignores an operator in the error state', () => {
    const error = errorState();

    expect(run(error, op('add'))).toBe(error);
  });
});

describe('unary operator', () => {
  it('sends sqrt for the first operand', () => {
    expect(run(typed('144'), SQRT)).toMatchObject({
      status: 'waiting',
      pending: { id: 1, operation: 'sqrt', operands: ['144'], then: null },
    });
  });

  it('does nothing when nothing was typed', () => {
    expect(run(initialState, SQRT)).toBe(initialState);
  });

  it('sends sqrt for the result when it follows a result', () => {
    expect(run(resultState('16'), SQRT)).toMatchObject({
      status: 'waiting',
      pending: { operation: 'sqrt', operands: ['16'] },
    });
  });

  it('drops a dangling decimal point before sending', () => {
    expect(run(typed('9.'), SQRT)).toMatchObject({ pending: { operands: ['9'] } });
  });

  it.each([
    ['operator chosen', run(typed('1'), op('add'))],
    ['entering the second operand', SECOND_OPERAND_STATE],
    ['waiting', waitingState()],
    ['error', errorState()],
  ])('is ignored in %s', (_label, state) => {
    expect(run(state, SQRT)).toBe(state);
  });
});

describe('equals', () => {
  it('sends both operands and waits for the result', () => {
    expect(run(SECOND_OPERAND_STATE, EQUALS)).toMatchObject({
      status: 'waiting',
      lastRequestId: 1,
      pending: { id: 1, operation: 'add', operands: ['12', '7'], then: null },
    });
  });

  it('keeps decimal operands as typed strings', () => {
    const state = run(typed('12.50'), op('add'), ...keys('7.25'), EQUALS);

    expect(state).toMatchObject({ pending: { operands: ['12.50', '7.25'] } });
  });

  it('drops a dangling decimal point from the second operand', () => {
    expect(run(typed('1'), op('add'), ...keys('2.'), EQUALS)).toMatchObject({
      pending: { operands: ['1', '2'] },
    });
  });

  it('gives each request a new id', () => {
    const second = run(resultState(), op('add'), ...keys('1'), EQUALS);

    expect(second).toMatchObject({ pending: { id: 2 }, lastRequestId: 2 });
  });

  it.each([
    ['nothing typed', initialState],
    ['only the first operand typed', typed('12')],
    ['an operator with no second operand', run(typed('12'), op('add'))],
    ['a result', resultState()],
    ['a request pending', waitingState()],
    ['an error', errorState()],
  ])('does nothing with %s', (_label, state) => {
    expect(run(state, EQUALS)).toBe(state);
  });
});

describe('backspace', () => {
  it('removes the last character of the first operand', () => {
    expect(run(typed('123'), BACKSPACE)).toMatchObject({ operand: '12' });
  });

  it('does nothing on an empty first operand', () => {
    expect(run(initialState, BACKSPACE)).toBe(initialState);
  });

  it('removes a negative sign together with the last digit', () => {
    expect(run(typed('5'), TOGGLE, BACKSPACE)).toMatchObject({ operand: '' });
  });

  it('removes the operator and goes back to the first operand', () => {
    expect(run(typed('12'), op('add'), BACKSPACE)).toMatchObject({
      status: 'entering_first',
      operand: '12',
    });
  });

  it('removes the last character of the second operand', () => {
    expect(run(SECOND_OPERAND_STATE, ...keys('3'), BACKSPACE)).toMatchObject({ operand: '7' });
  });

  it('returns to operator chosen when the second operand becomes empty', () => {
    expect(run(SECOND_OPERAND_STATE, BACKSPACE)).toMatchObject({
      status: 'operator_chosen',
      left: '12',
      operator: 'add',
    });
  });

  it('clears a result back to the start', () => {
    expect(run(resultState(), BACKSPACE)).toMatchObject({ status: 'entering_first', operand: '' });
  });

  it('clears an error back to the start', () => {
    expect(run(errorState(), BACKSPACE)).toMatchObject({ status: 'entering_first', operand: '' });
  });

  it('does nothing while a request is pending', () => {
    const waiting = waitingState();

    expect(run(waiting, BACKSPACE)).toBe(waiting);
  });
});

describe('clear', () => {
  it.each([
    ['entering the first operand', typed('12')],
    ['operator chosen', run(typed('1'), op('add'))],
    ['entering the second operand', SECOND_OPERAND_STATE],
    ['waiting', waitingState()],
    ['showing a result', resultState()],
    ['error', errorState()],
  ])('returns to the start from %s', (_label, state) => {
    expect(run(state, CLEAR)).toMatchObject({ status: 'entering_first', operand: '' });
  });

  it('keeps the request counter so ids never repeat', () => {
    expect(run(waitingState(), CLEAR)).toMatchObject({ lastRequestId: 1 });
    expect(run(waitingState(), CLEAR, ...keys('1'), op('add'), ...keys('1'), EQUALS)).toMatchObject(
      {
        pending: { id: 2 },
      },
    );
  });

  it('makes the answer to a cleared request stale', () => {
    const cleared = run(waitingState(), CLEAR);

    expect(run(cleared, { type: 'result_received', requestId: 1, result: '19' })).toBe(cleared);
  });
});

describe('toggle sign', () => {
  it('makes the first operand negative and back', () => {
    const negative = run(typed('5'), TOGGLE);

    expect(negative).toMatchObject({ operand: '-5' });
    expect(run(negative, TOGGLE)).toMatchObject({ operand: '5' });
  });

  it('keeps 0 as 0', () => {
    const state = typed('0');

    expect(run(state, TOGGLE)).toBe(state);
  });

  it('keeps 0.0 unsigned', () => {
    const state = typed('0.0');

    expect(run(state, TOGGLE)).toBe(state);
  });

  it('keeps an empty operand empty', () => {
    expect(run(initialState, TOGGLE)).toBe(initialState);
  });

  it('negates the second operand', () => {
    expect(run(SECOND_OPERAND_STATE, TOGGLE)).toMatchObject({ operand: '-7' });
  });

  it('turns a result into a negative first operand', () => {
    expect(run(resultState('2.5'), TOGGLE)).toMatchObject({
      status: 'entering_first',
      operand: '-2.5',
    });
  });

  it('keeps a result of 0 as it is', () => {
    const zero = resultState('0');

    expect(run(zero, TOGGLE)).toBe(zero);
  });

  it('is ignored with an operator chosen and no second operand', () => {
    const state = run(typed('1'), op('add'));

    expect(run(state, TOGGLE)).toBe(state);
  });

  it('is ignored while a request is pending', () => {
    const waiting = waitingState();

    expect(run(waiting, TOGGLE)).toBe(waiting);
  });
});

describe('result received', () => {
  it('shows the result and the request that produced it', () => {
    expect(run(waitingState(), { type: 'result_received', requestId: 1, result: '19' })).toEqual({
      status: 'showing_result',
      lastRequestId: 1,
      result: '19',
      request: { operation: 'add', operands: ['12', '7'] },
    });
  });

  it('keeps the result as the exact string from the server', () => {
    const result = '12345678901234567890.123456789012345678901';

    expect(run(waitingState(), { type: 'result_received', requestId: 1, result })).toMatchObject({
      result,
    });
  });

  it('ignores an answer to an older request', () => {
    const second = run(waitingState(), CLEAR, ...keys('1'), op('add'), ...keys('1'), EQUALS);

    expect(run(second, { type: 'result_received', requestId: 1, result: '19' })).toBe(second);
  });

  it('ignores an answer when no request is pending', () => {
    const state = typed('12');

    expect(run(state, { type: 'result_received', requestId: 1, result: '19' })).toBe(state);
  });
});

describe('request failed', () => {
  it('moves to the error state and keeps the request for a retry', () => {
    expect(
      run(waitingState(), { type: 'request_failed', requestId: 1, error: API_ERROR }),
    ).toMatchObject({
      status: 'error',
      error: API_ERROR,
      pending: { id: 1, operation: 'add', operands: ['12', '7'] },
    });
  });

  it('ignores a failure of an older request', () => {
    const second = run(waitingState(), CLEAR, ...keys('1'), op('add'), ...keys('1'), EQUALS);

    expect(run(second, { type: 'request_failed', requestId: 1, error: API_ERROR })).toBe(second);
  });

  it('ignores a failure when no request is pending', () => {
    const state = typed('12');

    expect(run(state, { type: 'request_failed', requestId: 1, error: API_ERROR })).toBe(state);
  });
});

describe('retry', () => {
  it('sends the same request again under a new id', () => {
    expect(run(errorState(NETWORK_ERROR), { type: 'retry' })).toMatchObject({
      status: 'waiting',
      lastRequestId: 2,
      pending: { id: 2, operation: 'add', operands: ['12', '7'], then: null },
    });
  });

  it('keeps the queued operator of a chained request', () => {
    const failed = run(SECOND_OPERAND_STATE, op('multiply'), {
      type: 'request_failed',
      requestId: 1,
      error: NETWORK_ERROR,
    });

    expect(run(failed, { type: 'retry' })).toMatchObject({ pending: { then: 'multiply' } });
  });

  it('does nothing for an error a retry cannot fix', () => {
    const error = errorState(API_ERROR);

    expect(run(error, { type: 'retry' })).toBe(error);
  });

  it('does nothing outside the error state', () => {
    const state = typed('12');

    expect(run(state, { type: 'retry' })).toBe(state);
  });
});

describe('reuse from history', () => {
  const REUSE: CalculatorAction = { type: 'reuse', result: '42.5' };

  it('shows the reused result with no expression', () => {
    expect(run(initialState, REUSE)).toMatchObject({
      status: 'showing_result',
      result: '42.5',
      request: null,
    });
  });

  it('replaces a result being shown', () => {
    expect(run(resultState('1'), REUSE)).toMatchObject({ result: '42.5', request: null });
  });

  it('replaces an error', () => {
    expect(run(errorState(), REUSE)).toMatchObject({ status: 'showing_result', result: '42.5' });
  });

  it('replaces a half-typed first operand', () => {
    expect(run(typed('12'), REUSE)).toMatchObject({ status: 'showing_result', result: '42.5' });
  });

  it('becomes the second operand when an operator is chosen', () => {
    expect(run(typed('1'), op('add'), REUSE)).toMatchObject({
      status: 'entering_second',
      left: '1',
      operand: '42.5',
    });
  });

  it('replaces the second operand being typed', () => {
    expect(run(SECOND_OPERAND_STATE, REUSE)).toMatchObject({ operand: '42.5' });
  });

  it('can be the first operand of the next calculation', () => {
    expect(run(initialState, REUSE, op('multiply'), ...keys('2'), EQUALS)).toMatchObject({
      pending: { operation: 'multiply', operands: ['42.5', '2'] },
    });
  });

  it('is ignored while a request is pending', () => {
    const waiting = waitingState();

    expect(run(waiting, REUSE)).toBe(waiting);
  });
});

describe('a result longer than an operand may be', () => {
  const LONG = '9'.repeat(MAX_OPERAND_LENGTH + 1);
  const FITS = '9'.repeat(MAX_OPERAND_LENGTH);
  const NOTICE = { code: 'RESULT_TOO_LONG_TO_REUSE', retryable: false };

  it('stays on screen with a notice instead of becoming the next left operand (chain)', () => {
    const chained = run(typed('12'), op('add'), ...keys('7'), op('multiply'));

    expect(run(chained, { type: 'result_received', requestId: 1, result: LONG })).toMatchObject({
      status: 'showing_result',
      result: LONG,
      request: { operation: 'add', operands: ['12', '7'] },
      notice: NOTICE,
    });
  });

  it('is still chained when it fits exactly', () => {
    const chained = run(typed('12'), op('add'), ...keys('7'), op('multiply'));

    expect(run(chained, { type: 'result_received', requestId: 1, result: FITS })).toMatchObject({
      status: 'operator_chosen',
      left: FITS,
    });
  });

  it('refuses an operator pressed on it', () => {
    const shown = resultState(LONG);

    expect(run(shown, op('add'))).toMatchObject({
      status: 'showing_result',
      result: LONG,
      notice: NOTICE,
    });
  });

  it('refuses a square root of it', () => {
    expect(run(resultState(LONG), SQRT)).toMatchObject({
      status: 'showing_result',
      result: LONG,
      notice: NOTICE,
    });
  });

  it('refuses a sign flip of it', () => {
    expect(run(resultState(LONG), TOGGLE)).toMatchObject({
      status: 'showing_result',
      notice: NOTICE,
    });
  });

  it('refuses a sign flip when the minus sign would not fit', () => {
    expect(run(resultState(FITS), TOGGLE)).toMatchObject({
      status: 'showing_result',
      result: FITS,
      notice: NOTICE,
    });
  });

  it('flips the sign of a negative result that fits exactly', () => {
    const negative = `-${'9'.repeat(MAX_OPERAND_LENGTH - 1)}`;

    expect(run(resultState(negative), TOGGLE)).toMatchObject({
      status: 'entering_first',
      operand: negative.slice(1),
    });
  });

  it('is shown, with a notice, when reused from history', () => {
    expect(run(initialState, { type: 'reuse', result: LONG })).toMatchObject({
      status: 'showing_result',
      result: LONG,
      request: null,
      notice: NOTICE,
    });
  });

  it('is not pushed into a calculation in progress when reused from history', () => {
    const second = run(typed('5'), op('add'), ...keys('2'));

    expect(run(second, { type: 'reuse', result: LONG })).toMatchObject({
      status: 'showing_result',
      result: LONG,
      notice: NOTICE,
    });
  });

  it('lets digits start a fresh calculation, which drops the notice', () => {
    expect(run(resultState(LONG), op('add'), ...keys('4'))).toEqual({
      status: 'entering_first',
      lastRequestId: 1,
      operand: '4',
    });
  });

  it('has no notice when it fits', () => {
    expect(resultState(FITS)).not.toHaveProperty('notice');
  });
});

describe('createInitialState', () => {
  it('carries the request counter forward', () => {
    expect(createInitialState(7)).toEqual({
      status: 'entering_first',
      operand: '',
      lastRequestId: 7,
    });
  });
});

describe('createCalculatorReducer', () => {
  const SHORT = 5;
  const shortReducer = createCalculatorReducer({ maxOperandLength: SHORT });
  const runShort = (...actions: CalculatorAction[]) =>
    actions.reduce<CalculatorState>(shortReducer, initialState);

  it('stops typing at the limit it was built with, not at the built-in one', () => {
    const state = runShort(...keys('1234567'));

    expect(state).toMatchObject({ operand: '12345' });
  });

  it('refuses a decimal point that would leave no room for a digit after it', () => {
    expect(runShort(...keys('1234.'))).toMatchObject({ operand: '1234' });
    expect(runShort(...keys('123.'))).toMatchObject({ operand: '123.' });
  });

  it('counts the minus sign when negating', () => {
    expect(runShort(...keys('12345'), TOGGLE)).toMatchObject({ operand: '12345' });
    expect(runShort(...keys('1234'), TOGGLE)).toMatchObject({ operand: '-1234' });
  });

  it('treats a result longer than its limit as too long to reuse', () => {
    const showing: CalculatorState = {
      status: 'showing_result',
      lastRequestId: 1,
      result: '123456',
      request: null,
    };

    const state = shortReducer(showing, { type: 'operator', operator: 'add' });

    expect(state).toMatchObject({
      status: 'showing_result',
      notice: { code: 'RESULT_TOO_LONG_TO_REUSE' },
    });
  });

  it('is the same reducer as calculatorReducer when given the built-in limit', () => {
    const builtIn = createCalculatorReducer({ maxOperandLength: MAX_OPERAND_LENGTH });
    const action: CalculatorAction = { type: 'digit', digit: '7' };

    expect(builtIn(initialState, action)).toEqual(calculatorReducer(initialState, action));
  });
});
