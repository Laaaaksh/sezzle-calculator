import { describe, expect, it } from 'vitest';
import { activeOperatorOf, describeDisplay } from './display';
import type { CalculatorError, CalculatorState, PendingRequest } from './state';

const PENDING: PendingRequest = { id: 1, operation: 'add', operands: ['12.5', '7.25'], then: null };
const ERROR: CalculatorError = {
  code: 'NETWORK_ERROR',
  message: 'x',
  requestId: undefined,
  retryable: true,
};

type WithoutRequestId<T> = T extends unknown ? Omit<T, 'lastRequestId'> : never;

function describeState(state: WithoutRequestId<CalculatorState>) {
  return describeDisplay({ ...state, lastRequestId: 1 });
}

describe('describeDisplay', () => {
  it('shows 0 before anything is typed', () => {
    expect(describeState({ status: 'entering_first', operand: '' })).toEqual({
      expression: '',
      value: '0',
    });
  });

  it('shows the operand being typed', () => {
    expect(describeState({ status: 'entering_first', operand: '12.' }).value).toBe('12.');
  });

  it('shows the first operand and the operator once an operator is chosen', () => {
    expect(describeState({ status: 'operator_chosen', left: '12', operator: 'multiply' })).toEqual({
      expression: '12 ×',
      value: '12',
    });
  });

  it('shows the expression so far above the second operand', () => {
    expect(
      describeState({ status: 'entering_second', left: '12', operator: 'add', operand: '7' }),
    ).toEqual({ expression: '12 +', value: '7' });
  });

  it('keeps the request and the last operand on screen while waiting', () => {
    expect(describeState({ status: 'waiting', pending: PENDING })).toEqual({
      expression: '12.5 + 7.25 =',
      value: '7.25',
    });
  });

  it('falls back to 0 when the pending request has no operands', () => {
    const state = { status: 'waiting', pending: { ...PENDING, operands: [] } } as const;

    expect(describeState(state).value).toBe('0');
  });

  it('shows the request without "=" and the word Error after an error', () => {
    expect(describeState({ status: 'error', error: ERROR, pending: PENDING })).toEqual({
      expression: '12.5 + 7.25',
      value: 'Error',
    });
  });

  it('shows the expression and the result', () => {
    expect(
      describeState({
        status: 'showing_result',
        result: '19.75',
        request: { operation: 'add', operands: ['12.5', '7.25'] },
      }),
    ).toEqual({ expression: '12.5 + 7.25 =', value: '19.75' });
  });

  it('shows only the result when it was reused from history', () => {
    expect(describeState({ status: 'showing_result', result: '19.75', request: null })).toEqual({
      expression: '',
      value: '19.75',
    });
  });
});

describe('activeOperatorOf', () => {
  it('is the operator once it is chosen', () => {
    expect(
      activeOperatorOf({ status: 'operator_chosen', left: '1', operator: 'add', lastRequestId: 0 }),
    ).toBe('add');
  });

  it('stays the operator while the second operand is typed', () => {
    expect(
      activeOperatorOf({
        status: 'entering_second',
        left: '1',
        operator: 'divide',
        operand: '2',
        lastRequestId: 0,
      }),
    ).toBe('divide');
  });

  it.each<CalculatorState>([
    { status: 'entering_first', operand: '1', lastRequestId: 0 },
    { status: 'showing_result', result: '2', request: null, lastRequestId: 0 },
  ])('is null in the $status state', (state) => {
    expect(activeOperatorOf(state)).toBeNull();
  });
});
