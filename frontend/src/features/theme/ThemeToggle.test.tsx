import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { THEME_ATTRIBUTE, THEME_STORAGE_KEY } from './theme';
import { ThemeToggle } from './ThemeToggle';

afterEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute(THEME_ATTRIBUTE);
  vi.restoreAllMocks();
});

function option(name: string) {
  return screen.getByRole('button', { name });
}

describe('ThemeToggle', () => {
  it('is a labelled group of three options', () => {
    render(<ThemeToggle />);

    const group = screen.getByRole('group', { name: 'Theme' });

    expect(group).toContainElement(option('System'));
    expect(group).toContainElement(option('Light'));
    expect(group).toContainElement(option('Dark'));
  });

  it('defaults to system, announced as pressed, with no data-theme', () => {
    render(<ThemeToggle />);

    expect(option('System')).toHaveAttribute('aria-pressed', 'true');
    expect(option('Dark')).toHaveAttribute('aria-pressed', 'false');
    expect(document.documentElement).not.toHaveAttribute(THEME_ATTRIBUTE);
  });

  it('applies and saves the chosen theme', async () => {
    const user = userEvent.setup();
    render(<ThemeToggle />);

    await user.click(option('Dark'));

    expect(option('Dark')).toHaveAttribute('aria-pressed', 'true');
    expect(option('System')).toHaveAttribute('aria-pressed', 'false');
    expect(document.documentElement).toHaveAttribute(THEME_ATTRIBUTE, 'dark');
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark');
  });

  it('returns to system and clears data-theme', async () => {
    const user = userEvent.setup();
    render(<ThemeToggle />);
    await user.click(option('Light'));

    await user.click(option('System'));

    expect(document.documentElement).not.toHaveAttribute(THEME_ATTRIBUTE);
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('system');
  });

  it('restores a saved choice on load', () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, 'light');

    render(<ThemeToggle />);

    expect(option('Light')).toHaveAttribute('aria-pressed', 'true');
    expect(document.documentElement).toHaveAttribute(THEME_ATTRIBUTE, 'light');
  });

  it('still switches theme when storage is blocked', async () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    const user = userEvent.setup();
    render(<ThemeToggle />);

    await user.click(option('Dark'));

    expect(document.documentElement).toHaveAttribute(THEME_ATTRIBUTE, 'dark');
  });
});
