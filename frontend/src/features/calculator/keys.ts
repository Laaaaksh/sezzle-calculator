import type { OperationType } from '../../api';
import type { Digit } from '../../constants/calculator';
import type { CalculatorAction } from './state';

export type KeyKind = 'digit' | 'operator' | 'function' | 'equals';

export interface KeyDefinition {
  id: string;
  label: string;
  /** Required for every symbol key, because a screen reader cannot say "÷" or "⌫" well. */
  ariaLabel?: string;
  action: CalculatorAction;
  kind: KeyKind;
  wide?: boolean;
  /** Extra help shown as a tooltip and read after the name, for a key whose meaning is not obvious. */
  hint?: string;
}

const digitKey = (digit: Digit): KeyDefinition => ({
  id: `digit-${digit}`,
  label: digit,
  action: { type: 'digit', digit },
  kind: 'digit',
});

// Written out key by key, not built from DIGITS with a cast: the Record type makes a missing
// digit a compile error instead of an undefined key at runtime.
const digitKeys: Record<Digit, KeyDefinition> = {
  '0': digitKey('0'),
  '1': digitKey('1'),
  '2': digitKey('2'),
  '3': digitKey('3'),
  '4': digitKey('4'),
  '5': digitKey('5'),
  '6': digitKey('6'),
  '7': digitKey('7'),
  '8': digitKey('8'),
  '9': digitKey('9'),
};

const symbol = (
  id: string,
  label: string,
  ariaLabel: string,
  action: CalculatorAction,
  kind: KeyKind,
  wide = false,
  hint?: string,
): KeyDefinition => ({
  id,
  label,
  ariaLabel,
  action,
  kind,
  ...(wide ? { wide } : {}),
  ...(hint === undefined ? {} : { hint }),
});

/**
 * The % key means "x percent of y", not "divide by 100". The visible label says "of" and the hint
 * spells out the order, because a bare % reads differently on every phone calculator.
 */
export const PERCENT_KEY_LABEL = '% of';
export const PERCENT_KEY_ARIA_LABEL = 'Percent of';
export const PERCENT_KEY_HINT = 'x percent of y: type x, press % of, then type y';

/** In reading order for a four-column grid: a double-width row, then four columns by five rows. */
export const KEYS: readonly KeyDefinition[] = [
  symbol('sqrt', '√', 'Square root', { type: 'unary', operation: 'sqrt' }, 'function', true),
  symbol('power', '^', 'Power', { type: 'operator', operator: 'power' }, 'operator', true),
  symbol('clear', 'C', 'Clear', { type: 'clear' }, 'function'),
  symbol('backspace', '⌫', 'Backspace', { type: 'backspace' }, 'function'),
  symbol(
    'percentage',
    PERCENT_KEY_LABEL,
    PERCENT_KEY_ARIA_LABEL,
    { type: 'operator', operator: 'percentage' },
    'operator',
    false,
    PERCENT_KEY_HINT,
  ),
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

export function operationForAction(action: CalculatorAction): OperationType | null {
  switch (action.type) {
    case 'operator':
      return action.operator;
    case 'unary':
      return action.operation;
    default:
      return null;
  }
}

/**
 * False only when the server reported its operations and this action's operation is not among
 * them. `null` means the list is unknown (still loading, or the fetch failed), and then every
 * key stays usable because the server validates each request anyway.
 */
export function isActionAvailable(
  action: CalculatorAction,
  supported: ReadonlySet<string> | null,
): boolean {
  const operation = operationForAction(action);
  return supported === null || operation === null || supported.has(operation);
}
