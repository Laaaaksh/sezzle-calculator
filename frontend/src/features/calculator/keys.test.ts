import { describe, expect, it } from 'vitest';
import { DIGITS } from '../../constants/calculator';
import { isActionAvailable, KEYS, operationForAction } from './keys';

describe('KEYS', () => {
  it('has an aria-label on every key that is not a plain digit', () => {
    const missing = KEYS.filter(
      (key) => !DIGITS.some((digit) => digit === key.label) && key.ariaLabel === undefined,
    );

    expect(missing).toEqual([]);
  });

  it('gives every key a unique id', () => {
    expect(new Set(KEYS.map((key) => key.id)).size).toBe(KEYS.length);
  });

  it('gives every key a unique aria-label', () => {
    const labels = KEYS.flatMap((key) => (key.ariaLabel === undefined ? [] : [key.ariaLabel]));

    expect(new Set(labels).size).toBe(labels.length);
  });

  it('has all ten digits', () => {
    const digits = KEYS.flatMap((key) => (key.action.type === 'digit' ? [key.action.digit] : []));

    expect([...digits].sort()).toEqual([...DIGITS]);
  });

  it.each(['add', 'subtract', 'multiply', 'divide', 'power', 'percentage'])(
    'has an operator key for %s',
    (operator) => {
      expect(
        KEYS.some((key) => key.action.type === 'operator' && key.action.operator === operator),
      ).toBe(true);
    },
  );

  it.each(['unary', 'equals', 'backspace', 'clear', 'toggle_sign', 'decimal'])(
    'has a key for the %s action',
    (type) => {
      expect(KEYS.some((key) => key.action.type === type)).toBe(true);
    },
  );
});

describe('operationForAction', () => {
  it('names the operation of an operator and of a unary action', () => {
    expect(operationForAction({ type: 'operator', operator: 'power' })).toBe('power');
    expect(operationForAction({ type: 'unary', operation: 'sqrt' })).toBe('sqrt');
  });

  it('is null for actions that never call the server', () => {
    expect(operationForAction({ type: 'clear' })).toBeNull();
    expect(operationForAction({ type: 'equals' })).toBeNull();
  });
});

describe('isActionAvailable', () => {
  const supported = new Set(['add'] as const);

  it('allows everything while the list is unknown', () => {
    expect(isActionAvailable({ type: 'unary', operation: 'sqrt' }, null)).toBe(true);
  });

  it('blocks an operation the server did not report', () => {
    expect(isActionAvailable({ type: 'operator', operator: 'power' }, supported)).toBe(false);
  });

  it('allows a reported operation and every action without one', () => {
    expect(isActionAvailable({ type: 'operator', operator: 'add' }, supported)).toBe(true);
    expect(isActionAvailable({ type: 'digit', digit: '1' }, supported)).toBe(true);
  });
});
