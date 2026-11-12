import { describe, expect, it } from 'vitest';
import { DIGITS } from '../../constants/calculator';
import { KEYS } from './keys';

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
