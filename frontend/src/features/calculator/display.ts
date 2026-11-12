import { OPERATION_SYMBOLS } from '../../constants/calculator';
import { formatExpression } from './expression';
import type { CalculatorState } from './state';

export interface DisplayLines {
  /** Small line above the value: what is being calculated. Empty when there is nothing yet. */
  expression: string;
  /** Big line: what is being typed, or the result. */
  value: string;
}

const ZERO = '0';
const EQUALS_SIGN = '=';

/** What the display shows for a state. Pure, so each state is easy to test. */
export function describeDisplay(state: CalculatorState): DisplayLines {
  switch (state.status) {
    case 'entering_first':
      return { expression: '', value: state.operand === '' ? ZERO : state.operand };
    case 'operator_chosen':
      return {
        expression: `${state.left} ${OPERATION_SYMBOLS[state.operator]}`,
        value: state.left,
      };
    case 'entering_second':
      return {
        expression: `${state.left} ${OPERATION_SYMBOLS[state.operator]}`,
        value: state.operand,
      };
    case 'waiting':
    case 'error':
      // While waiting or after a failure the last operand stays up, so the screen does not jump.
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
