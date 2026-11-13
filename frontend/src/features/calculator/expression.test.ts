import { describe, expect, it } from 'vitest';
import { formatExpression, formatPendingExpression } from './expression';

describe('formatExpression', () => {
  it('writes a binary operation with its symbol', () => {
    expect(formatExpression({ operation: 'add', operands: ['12.5', '7.25'] })).toBe('12.5 + 7.25');
  });

  it('uses display symbols for multiply and divide', () => {
    expect(formatExpression({ operation: 'multiply', operands: ['3', '4'] })).toBe('3 × 4');
    expect(formatExpression({ operation: 'divide', operands: ['8', '2'] })).toBe('8 ÷ 2');
  });

  it('writes sqrt as a function', () => {
    expect(formatExpression({ operation: 'sqrt', operands: ['144'] })).toBe('√(144)');
  });

  it('writes percentage as "x% of y"', () => {
    expect(formatExpression({ operation: 'percentage', operands: ['20', '50'] })).toBe('20% of 50');
  });
});

describe('formatPendingExpression', () => {
  it('writes an operator after the first operand', () => {
    expect(formatPendingExpression('12', 'multiply')).toBe('12 ×');
  });

  it('writes percentage as "x% of" so the screen says what the key will do', () => {
    expect(formatPendingExpression('20', 'percentage')).toBe('20% of');
  });
});
