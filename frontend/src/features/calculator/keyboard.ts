import { DIGITS, type BinaryOperation } from '../../constants/calculator';
import type { CalculatorAction } from './state';

// A Map, not an object, so a key named like an Object method can never match by accident.
const OPERATOR_KEYS = new Map<string, BinaryOperation>([
  ['+', 'add'],
  ['-', 'subtract'],
  ['*', 'multiply'],
  ['x', 'multiply'],
  ['/', 'divide'],
  ['^', 'power'],
  ['%', 'percentage'],
]);

const EQUALS_KEYS = ['Enter', '='];

/**
 * What a key press means for the calculator, or null when the key is not a shortcut.
 * Letters are matched without regard to case so Caps Lock does not break x and r.
 */
export function actionForKey(key: string): CalculatorAction | null {
  const digit = DIGITS.find((candidate) => candidate === key);
  if (digit !== undefined) {
    return { type: 'digit', digit };
  }
  const operator = OPERATOR_KEYS.get(key.toLowerCase());
  if (operator !== undefined) {
    return { type: 'operator', operator };
  }
  if (EQUALS_KEYS.includes(key)) {
    return { type: 'equals' };
  }
  switch (key.toLowerCase()) {
    case '.':
      return { type: 'decimal' };
    case 'backspace':
      return { type: 'backspace' };
    case 'escape':
      return { type: 'clear' };
    case 'r':
      return { type: 'unary', operation: 'sqrt' };
    default:
      return null;
  }
}
