import { OPERATION_SYMBOLS } from '../../constants/calculator';
import type { BinaryOperation } from '../../constants/calculator';
import type { CalculationRequest } from './state';

/**
 * The expression so far, while the second operand is still to come: "12 +", "20% of". Percentage
 * reads "20% of", never a bare "20 %", so the screen says what the % key will do.
 */
export function formatPendingExpression(left: string, operator: BinaryOperation): string {
  return operator === 'percentage'
    ? `${left}${OPERATION_SYMBOLS.percentage} of`
    : `${left} ${OPERATION_SYMBOLS[operator]}`;
}

/** Writes a request the way a person would read it: "12.5 + 7.25", "20% of 50", "√(144)". */
export function formatExpression({ operation, operands }: CalculationRequest): string {
  const [first = '', second = ''] = operands;
  switch (operation) {
    case 'sqrt':
      return `${OPERATION_SYMBOLS.sqrt}(${first})`;
    case 'percentage':
      return `${first}${OPERATION_SYMBOLS.percentage} of ${second}`;
    default:
      return `${first} ${OPERATION_SYMBOLS[operation]} ${second}`;
  }
}
