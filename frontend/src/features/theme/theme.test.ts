import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  applyTheme,
  isThemePreference,
  readStoredTheme,
  storeTheme,
  THEME_ATTRIBUTE,
  THEME_STORAGE_KEY,
} from './theme';

afterEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute(THEME_ATTRIBUTE);
  vi.restoreAllMocks();
});

describe('isThemePreference', () => {
  it.each(['system', 'light', 'dark'])('accepts %s', (value) => {
    expect(isThemePreference(value)).toBe(true);
  });

  it.each(['', 'Dark', 'auto', null, undefined, 1])('rejects %s', (value) => {
    expect(isThemePreference(value)).toBe(false);
  });
});

describe('readStoredTheme', () => {
  it('defaults to system when nothing is saved', () => {
    expect(readStoredTheme()).toBe('system');
  });

  it('returns a saved choice', () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, 'dark');

    expect(readStoredTheme()).toBe('dark');
  });

  it('ignores a saved value it does not recognise', () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, 'purple');

    expect(readStoredTheme()).toBe('system');
  });

  it('falls back to system when storage throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });

    expect(readStoredTheme()).toBe('system');
  });
});

describe('storeTheme', () => {
  it('saves the choice', () => {
    storeTheme('light');

    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('light');
  });

  it('does not throw when storage is blocked', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('full', 'QuotaExceededError');
    });

    expect(() => {
      storeTheme('dark');
    }).not.toThrow();
  });
});

describe('applyTheme', () => {
  it.each(['light', 'dark'] as const)('sets data-theme for %s', (preference) => {
    applyTheme(preference);

    expect(document.documentElement).toHaveAttribute(THEME_ATTRIBUTE, preference);
  });

  it('removes data-theme for system so the stylesheet follows the OS', () => {
    applyTheme('dark');

    applyTheme('system');

    expect(document.documentElement).not.toHaveAttribute(THEME_ATTRIBUTE);
  });
});
