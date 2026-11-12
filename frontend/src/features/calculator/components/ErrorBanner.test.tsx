import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { CalculatorError } from '../state';
import { ErrorBanner } from './ErrorBanner';

function error(overrides: Partial<CalculatorError>): CalculatorError {
  return {
    code: 'NETWORK_ERROR',
    requestId: undefined,
    retryable: true,
    ...overrides,
  };
}

describe('ErrorBanner', () => {
  it('shows the friendly message, not the raw one', () => {
    render(
      <ErrorBanner
        error={error({ code: 'DIVISION_BY_ZERO', retryable: false })}
        onRetry={vi.fn()}
      />,
    );

    expect(screen.getByText("You can't divide by zero.")).toBeInTheDocument();
    expect(screen.queryByText('raw')).not.toBeInTheDocument();
  });

  it('offers Retry for a retryable error and calls onRetry', async () => {
    const onRetry = vi.fn();
    render(<ErrorBanner error={error({})} onRetry={onRetry} />);

    await userEvent.click(screen.getByRole('button', { name: 'Retry' }));

    expect(onRetry).toHaveBeenCalledOnce();
  });

  it('has no Retry for an error a retry cannot fix', () => {
    render(<ErrorBanner error={error({ retryable: false })} onRetry={vi.fn()} />);

    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('announces the message as an alert, outside the polite status region', () => {
    render(<ErrorBanner error={error({})} onRetry={vi.fn()} />);

    expect(screen.getByRole('alert')).toHaveTextContent("Couldn't reach the server.");
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
