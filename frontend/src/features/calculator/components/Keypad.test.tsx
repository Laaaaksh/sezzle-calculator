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
});
