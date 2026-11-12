import { DIGITS } from '../../constants/calculator';
import type { CalculatorAction } from './state';

export type KeyKind = 'digit' | 'operator' | 'function' | 'equals';

export interface KeyDefinition {
  /** Unique, stable, used as the React key. */
  id: string;
  /** What the key shows. */
  label: string;
  /** Required for every symbol key, because a screen reader cannot say "÷" or "⌫" well. */
  ariaLabel?: string;
  action: CalculatorAction;
  /** Styling hook only. */
  kind: KeyKind;
  /** Takes two columns of the four-column grid. */
  wide?: boolean;
}

const digitKeys = Object.fromEntries(
  DIGITS.map((digit) => [
    digit,
    { id: `digit-${digit}`, label: digit, action: { type: 'digit', digit }, kind: 'digit' },
  ]),
) as Record<(typeof DIGITS)[number], KeyDefinition>;

const symbol = (
  id: string,
  label: string,
  ariaLabel: string,
  action: CalculatorAction,
  kind: KeyKind,
  wide = false,
): KeyDefinition => ({ id, label, ariaLabel, action, kind, ...(wide ? { wide } : {}) });

/** In reading order for a four-column grid: a double-width row, then four columns by five rows. */
export const KEYS: readonly KeyDefinition[] = [
  symbol('sqrt', '√', 'Square root', { type: 'unary', operation: 'sqrt' }, 'function', true),
  symbol('power', '^', 'Power', { type: 'operator', operator: 'power' }, 'operator', true),
  symbol('clear', 'C', 'Clear', { type: 'clear' }, 'function'),
  symbol('backspace', '⌫', 'Backspace', { type: 'backspace' }, 'function'),
  symbol('percentage', '%', 'Percent', { type: 'operator', operator: 'percentage' }, 'operator'),
  symbol('divide', '÷', 'Divide', { type: 'operator', operator: 'divide' }, 'operator'),
  digitKeys['7'],
  digitKeys['8'],
  digitKeys['9'],
  symbol('multiply', '×', 'Multiply', { type: 'operator', operator: 'multiply' }, 'operator'),
  digitKeys['4'],
  digitKeys['5'],
  digitKeys['6'],
  symbol('subtract', '-', 'Subtract', { type: 'operator', operator: 'subtract' }, 'operator'),
  digitKeys['1'],
  digitKeys['2'],
  digitKeys['3'],
  symbol('add', '+', 'Add', { type: 'operator', operator: 'add' }, 'operator'),
  symbol('toggle-sign', '±', 'Toggle sign', { type: 'toggle_sign' }, 'function'),
  digitKeys['0'],
  symbol('decimal', '.', 'Decimal point', { type: 'decimal' }, 'digit'),
  symbol('equals', '=', 'Equals', { type: 'equals' }, 'equals'),
];
