import { describe, expect, it } from 'vitest';
import { DIGITS } from '../../constants/calculator';
import { actionForKey } from './keyboard';
import type { CalculatorAction } from './state';

describe('actionForKey', () => {
  it.each(DIGITS)('maps %s to that digit', (digit) => {
    expect(actionForKey(digit)).toEqual({ type: 'digit', digit });
  });

  it('maps the dot to the decimal point', () => {
    expect(actionForKey('.')).toEqual({ type: 'decimal' });
  });

  it.each([
    ['+', 'add'],
    ['-', 'subtract'],
    ['*', 'multiply'],
    ['x', 'multiply'],
    ['X', 'multiply'],
    ['/', 'divide'],
    ['^', 'power'],
    ['%', 'percentage'],
  ] as const)('maps %s to %s', (key, operator) => {
    expect(actionForKey(key)).toEqual({ type: 'operator', operator });
  });

  it.each(['Enter', '='])('maps %s to equals', (key) => {
    expect(actionForKey(key)).toEqual({ type: 'equals' });
  });

  it('maps Backspace to backspace', () => {
    expect(actionForKey('Backspace')).toEqual({ type: 'backspace' });
  });

  it('maps Escape to clear', () => {
    expect(actionForKey('Escape')).toEqual({ type: 'clear' });
  });

  it.each(['r', 'R'])('maps %s to square root', (key) => {
    const expected: CalculatorAction = { type: 'unary', operation: 'sqrt' };

    expect(actionForKey(key)).toEqual(expected);
  });

  it.each(['a', 'Tab', 'Shift', 'ArrowLeft', ' ', ',', 'constructor'])(
    'maps %j to nothing',
    (key) => {
      expect(actionForKey(key)).toBeNull();
    },
  );
});
