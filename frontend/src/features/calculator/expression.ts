import { OPERATION_SYMBOLS } from '../../constants/calculator';
import type { CalculationRequest } from './state';

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
