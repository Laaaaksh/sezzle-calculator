import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { HistoryEntry } from '../history';
import { HistoryTape } from './HistoryTape';

const ENTRIES: HistoryEntry[] = [
  { id: 2, request: { operation: 'multiply', operands: ['3', '4'] }, result: '12' },
  { id: 1, request: { operation: 'add', operands: ['12.5', '7.25'] }, result: '19.75' },
];

describe('HistoryTape', () => {
  it('says so when there is nothing yet', () => {
    render(<HistoryTape entries={[]} onReuse={vi.fn()} />);

    expect(screen.getByText('No calculations yet.')).toBeInTheDocument();
  });

  it('lists each entry as an expression and its result, in the order given', () => {
    render(<HistoryTape entries={ENTRIES} onReuse={vi.fn()} />);

    const items = within(screen.getByRole('list')).getAllByRole('listitem');
    expect(items[0]).toHaveTextContent('3 × 4 = 12');
    expect(items[1]).toHaveTextContent('12.5 + 7.25 = 19.75');
  });

  it('calls onReuse with the entry that was clicked', async () => {
    const onReuse = vi.fn();
    render(<HistoryTape entries={ENTRIES} onReuse={onReuse} />);

    await userEvent.click(screen.getByRole('button', { name: /12\.5 \+ 7\.25/ }));

    expect(onReuse).toHaveBeenCalledExactlyOnceWith(ENTRIES[1]);
  });
});

describe('HistoryTape heading', () => {
  it('labels the section with its heading', () => {
    render(<HistoryTape entries={[]} onReuse={vi.fn()} />);

    expect(screen.getByRole('region', { name: 'History' })).toBeInTheDocument();
  });

  it('gives each tape its own heading id, so two on a page do not clash', () => {
    render(
      <>
        <HistoryTape entries={[]} onReuse={vi.fn()} />
        <HistoryTape entries={[]} onReuse={vi.fn()} />
      </>,
    );

    const [first, second] = screen.getAllByRole('heading', { name: 'History' });
    expect(first?.id).not.toBe('');
    expect(first?.id).not.toBe(second?.id);
  });

  it('adds the class a parent passes to the heading', () => {
    render(<HistoryTape entries={[]} onReuse={vi.fn()} headingClassName="from-parent" />);

    expect(screen.getByRole('heading', { name: 'History' })).toHaveClass('from-parent');
  });
});
