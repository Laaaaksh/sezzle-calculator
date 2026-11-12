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

  it('shows the value', () => {
    render(<Display state={entering('5')} isPending={false} />);

    expect(screen.getByRole('status')).toHaveTextContent('5');
  });

  it('marks the region busy while a request is pending', () => {
    render(<Display state={entering('5')} isPending />);

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });

  it('renders children inside the live region', () => {
    render(
      <Display state={entering('5')} isPending={false}>
        <p>child</p>
      </Display>,
    );

    expect(screen.getByRole('status')).toContainElement(screen.getByText('child'));
  });

  it.each([
    ['12345', 'normal'],
    ['1'.repeat(13), 'long'],
    ['1'.repeat(25), 'very-long'],
  ])('sets a size hint for %s', (value, size) => {
    render(<Display state={entering(value)} isPending={false} />);

    expect(screen.getByText(value)).toHaveAttribute('data-size', size);
  });
});
