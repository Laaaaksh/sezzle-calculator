import type { BinaryOperation } from '../../constants/calculator';
import { formatExpression, formatPendingExpression } from './expression';
import type { CalculatorState } from './state';

export interface DisplayLines {
  expression: string;
  value: string;
}

const ZERO = '0';
const EQUALS_SIGN = '=';
export const ERROR_VALUE = 'Error';

export function describeDisplay(state: CalculatorState): DisplayLines {
  switch (state.status) {
    case 'entering_first':
      return { expression: '', value: state.operand === '' ? ZERO : state.operand };
    case 'operator_chosen':
      return {
        expression: formatPendingExpression(state.left, state.operator),
        value: state.left,
      };
    case 'entering_second':
      return {
        expression: formatPendingExpression(state.left, state.operator),
        value: state.operand,
      };
    case 'error':
      // No "=" and no number: "9 ÷ 0 = 0" would read as an answer. The alert below says why.
      return { expression: formatExpression(state.pending), value: ERROR_VALUE };
    case 'waiting':
      // While waiting the last operand stays up, so the screen does not jump.
      return {
        expression: `${formatExpression(state.pending)} ${EQUALS_SIGN}`,
        value: state.pending.operands.at(-1) ?? ZERO,
      };
    case 'showing_result':
      return {
        expression: state.request ? `${formatExpression(state.request)} ${EQUALS_SIGN}` : '',
        value: state.result,
      };
  }
}

export function activeOperatorOf(state: CalculatorState): BinaryOperation | null {
  return state.status === 'operator_chosen' || state.status === 'entering_second'
    ? state.operator
    : null;
}
