import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { HistoryEntry } from '../history';
import { HistoryDrawer } from './HistoryDrawer';

const ENTRIES: HistoryEntry[] = [
  { id: 1, request: { operation: 'add', operands: ['12.5', '7.25'] }, result: '19.75' },
];

function toggle() {
  return screen.getByRole('button', { name: 'History' });
}

describe('HistoryDrawer', () => {
  it('starts closed', () => {
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);

    expect(toggle()).toHaveAttribute('aria-expanded', 'false');
  });

  it('points the button at the panel it controls', () => {
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);

    const panelId = toggle().getAttribute('aria-controls');

    expect(panelId).toBeTruthy();
    expect(document.getElementById(panelId ?? '')).toContainElement(
      screen.getByRole('region', { name: 'History' }),
    );
  });

  it('opens and closes with the button', async () => {
    const user = userEvent.setup();
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);

    await user.click(toggle());
    expect(toggle()).toHaveAttribute('aria-expanded', 'true');

    await user.click(toggle());
    expect(toggle()).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes on Escape from inside the panel and returns focus to the button', async () => {
    const user = userEvent.setup();
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);
    await user.click(toggle());
    await user.tab();
    expect(screen.getByRole('button', { name: /12\.5 \+ 7\.25/ })).toHaveFocus();

    await user.keyboard('{Escape}');

    expect(toggle()).toHaveAttribute('aria-expanded', 'false');
    expect(toggle()).toHaveFocus();
  });

  it('closes on Escape while the button itself has focus', async () => {
    const user = userEvent.setup();
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);
    await user.click(toggle());
    toggle().focus();

    await user.keyboard('{Escape}');

    expect(toggle()).toHaveAttribute('aria-expanded', 'false');
  });

  it('leaves Escape alone when the drawer is closed, so the calculator shortcut still works', () => {
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true });

    toggle().dispatchEvent(event);

    expect(event.defaultPrevented).toBe(false);
  });

  it('marks Escape as handled when it closes the drawer', async () => {
    const user = userEvent.setup();
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);
    await user.click(toggle());
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true });

    toggle().dispatchEvent(event);

    expect(event.defaultPrevented).toBe(true);
  });

  it('reuses an entry, closes the drawer and returns focus to the button', async () => {
    const user = userEvent.setup();
    const onReuse = vi.fn();
    render(<HistoryDrawer entries={ENTRIES} onReuse={onReuse} />);
    await user.click(toggle());

    await user.click(screen.getByRole('button', { name: /12\.5 \+ 7\.25/ }));

    expect(onReuse).toHaveBeenCalledExactlyOnceWith(ENTRIES[0]);
    expect(toggle()).toHaveAttribute('aria-expanded', 'false');
    expect(toggle()).toHaveFocus();
  });

  it('reuses an entry without giving it focus when the drawer was never opened (wide layout)', async () => {
    const user = userEvent.setup();
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);
    const entry = screen.getByRole('button', { name: /12\.5 \+ 7\.25/ });

    await user.click(entry);

    // Entries never take focus on mousedown, so Enter still means equals.
    expect(entry).not.toHaveFocus();
  });

  it('keeps the History heading for screen readers inside the panel', () => {
    render(<HistoryDrawer entries={ENTRIES} onReuse={vi.fn()} />);

    expect(screen.getByRole('heading', { name: 'History' })).toBeInTheDocument();
  });
});
