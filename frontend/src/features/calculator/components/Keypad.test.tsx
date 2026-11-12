import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { KEYS } from '../keys';
import { Keypad } from './Keypad';

describe('Keypad', () => {
  it('renders a real button for every key', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    expect(screen.getAllByRole('button')).toHaveLength(KEYS.length);
  });

  it('names a symbol key by its aria-label', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Divide' })).toHaveTextContent('÷');
  });

  it('names the percent key "Percent of" and shows "% of"', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Percent of' })).toHaveTextContent('% of');
  });

  it('explains the percent key in a tooltip and in its accessible description', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    const key = screen.getByRole('button', { name: 'Percent of' });

    expect(key).toHaveAttribute('title', 'x percent of y: type x, press % of, then type y');
    expect(key).toHaveAccessibleDescription('x percent of y: type x, press % of, then type y');
  });

  it('gives the other keys no tooltip', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Divide' })).not.toHaveAttribute('title');
  });

  it('passes the key action to onPress', async () => {
    const onPress = vi.fn();
    render(<Keypad isPending={false} onPress={onPress} />);

    await userEvent.click(screen.getByRole('button', { name: '7' }));

    expect(onPress).toHaveBeenCalledExactlyOnceWith({ type: 'digit', digit: '7' });
  });

  it('disables equals while a request is pending', () => {
    render(<Keypad isPending onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Equals' })).toBeDisabled();
  });

  it('leaves the other keys enabled while a request is pending', () => {
    render(<Keypad isPending onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Clear' })).toBeEnabled();
  });

  it('does not take focus on a pointer press, so Enter still means equals', async () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    await userEvent.click(screen.getByRole('button', { name: '7' }));

    expect(screen.getByRole('button', { name: '7' })).not.toHaveFocus();
  });

  it('marks only the active operator as pressed', () => {
    render(<Keypad isPending={false} activeOperator="add" onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Subtract' })).toHaveAttribute(
      'aria-pressed',
      'false',
    );
  });

  it('leaves digits and functions without a pressed state', () => {
    render(<Keypad isPending={false} activeOperator="add" onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: '7' })).not.toHaveAttribute('aria-pressed');
    expect(screen.getByRole('button', { name: 'Clear' })).not.toHaveAttribute('aria-pressed');
  });

  it('presses no operator when none is active', () => {
    render(<Keypad isPending={false} onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('marks keys for operations the server did not report as unavailable, with a reason', () => {
    render(<Keypad isPending={false} supportedOperations={new Set(['add'])} onPress={vi.fn()} />);

    const power = screen.getByRole('button', { name: 'Power' });
    expect(power).toHaveAttribute('aria-disabled', 'true');
    expect(power).toHaveAttribute('title', 'Not available on this server');
    expect(screen.getByRole('button', { name: 'Add' })).not.toHaveAttribute('aria-disabled');
    expect(screen.getByRole('button', { name: 'Add' })).not.toHaveAttribute('aria-describedby');
    expect(screen.getByRole('button', { name: '7' })).not.toHaveAttribute('aria-disabled');
  });

  it('does not press an unavailable key, but leaves it focusable', async () => {
    const onPress = vi.fn();
    render(<Keypad isPending={false} supportedOperations={new Set(['add'])} onPress={onPress} />);
    const sqrt = screen.getByRole('button', { name: 'Square root' });

    await userEvent.click(sqrt);

    expect(onPress).not.toHaveBeenCalled();
    expect(sqrt).toBeEnabled();
  });
});
