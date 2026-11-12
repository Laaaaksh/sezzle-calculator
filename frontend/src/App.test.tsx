import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import App from './App';
import { KEYS } from './features/calculator/keys';
import { expectNoAxeViolations } from './test/axe';

describe('App', () => {
  it('renders the calculator heading', () => {
    render(<App />);

    expect(screen.getByRole('heading', { name: 'Calculator' })).toBeInTheDocument();
  });

  it('offers the theme toggle', () => {
    render(<App />);

    expect(screen.getByRole('group', { name: 'Theme' })).toBeInTheDocument();
  });

  it('has no axe violations', async () => {
    const { container } = render(<App />);

    await expectNoAxeViolations(container);
  });

  it('tabs through the page in reading order: theme, then keys, then history', async () => {
    const user = userEvent.setup();
    render(<App />);
    const expected = [
      'System',
      'Light',
      'Dark',
      ...KEYS.map((key) => key.ariaLabel ?? key.label),
      'History',
    ];

    for (const name of expected) {
      await user.tab();
      const focused = document.activeElement;
      expect(focused?.getAttribute('aria-label') ?? focused?.textContent).toBe(name);
    }
  });

  it('never reorders focus with a positive tabindex', () => {
    const { container } = render(<App />);

    const positive = [...container.querySelectorAll('[tabindex]')].filter(
      (element) => Number(element.getAttribute('tabindex')) > 0,
    );

    expect(positive).toEqual([]);
  });
});
