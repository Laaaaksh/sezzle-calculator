import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { CalculatorState } from '../state';
import { Display } from './Display';

function entering(operand: string): CalculatorState {
  return { status: 'entering_first', operand, lastRequestId: 0 };
}

describe('Display', () => {
  it('is a polite live status region', () => {
    render(<Display state={entering('5')} isPending={false} />);

    expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite');
  });

  it('shows the value in the live region', () => {
    render(<Display state={entering('5')} isPending={false} />);

    expect(screen.getByRole('status')).toHaveTextContent('5');
  });

  it('keeps the expression outside the live region so a keystroke announces only the value', () => {
    render(
      <Display
        state={{ status: 'operator_chosen', left: '12', operator: 'add', lastRequestId: 0 }}
        isPending={false}
      />,
    );

    expect(screen.getByRole('group', { name: 'Display' })).toHaveTextContent('12 +');
    expect(screen.getByRole('status')).toHaveTextContent(/^12$/);
  });

  it('can take focus from a script but is not in the tab order', () => {
    render(<Display state={entering('5')} isPending={false} />);
    const region = screen.getByRole('group', { name: 'Display' });

    region.focus();

    expect(region).toHaveFocus();
    expect(region).toHaveAttribute('tabindex', '-1');
  });

  it('marks the region busy while a request is pending', () => {
    render(<Display state={entering('5')} isPending />);

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });

  it.each([
    ['5', '1'],
    ['12345', '5'],
    ['1'.repeat(25), '25'],
  ])('passes the length of %s to the stylesheet so it can size the text', (value, length) => {
    render(<Display state={entering(value)} isPending={false} />);

    expect(screen.getByText(value).style.getPropertyValue('--value-length')).toBe(length);
  });

  it('never reports a zero length, which would divide by zero in the stylesheet', () => {
    render(
      <Display
        state={{
          status: 'entering_second',
          left: '1',
          operator: 'add',
          operand: '',
          lastRequestId: 0,
        }}
        isPending={false}
      />,
    );

    expect(
      screen.getAllByText('', { selector: 'p' }).at(-1)?.style.getPropertyValue('--value-length'),
    ).toBe('1');
  });

  it('restarts the entrance animation when a result arrives', () => {
    const { rerender } = render(<Display state={entering('2')} isPending={false} />);
    const typed = screen.getByText('2', { selector: 'p:last-of-type' });

    rerender(
      <Display
        state={{ status: 'showing_result', result: '4', request: null, lastRequestId: 1 }}
        isPending={false}
      />,
    );

    expect(typed).not.toBeInTheDocument();
    expect(screen.getByText('4')).toBeInTheDocument();
  });

  it('keeps the same element while digits are typed, so typing does not animate', () => {
    const { rerender } = render(<Display state={entering('2')} isPending={false} />);
    const typed = screen.getByText('2', { selector: 'p:last-of-type' });

    rerender(<Display state={entering('23')} isPending={false} />);

    expect(typed).toBeInTheDocument();
    expect(typed).toHaveTextContent('23');
  });

  it('shows Error without an equals sign in the error state', () => {
    render(
      <Display
        state={{
          status: 'error',
          lastRequestId: 1,
          error: { code: 'NETWORK_ERROR', message: 'x', requestId: undefined, retryable: true },
          pending: { id: 1, operation: 'divide', operands: ['9', '0'], then: null },
        }}
        isPending={false}
      />,
    );

    expect(screen.getByRole('group', { name: 'Display' })).toHaveTextContent('9 ÷ 0');
    expect(screen.getByRole('group', { name: 'Display' })).not.toHaveTextContent('=');
    expect(screen.getByText('Error')).toBeInTheDocument();
  });
});
