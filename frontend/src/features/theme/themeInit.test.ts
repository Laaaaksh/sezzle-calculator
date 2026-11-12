import { afterEach, describe, expect, it } from 'vitest';
import themeInitSource from '../../../public/theme-init.js?raw';
import { THEME_ATTRIBUTE, THEME_STORAGE_KEY } from './theme';

/** Runs the shipped script the way the browser does: as a classic script element. */
function runThemeInit() {
  const script = document.createElement('script');
  script.textContent = themeInitSource;
  document.head.append(script);
  script.remove();
}

afterEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute(THEME_ATTRIBUTE);
});

describe('public/theme-init.js', () => {
  it('uses the same storage key as the app, or a saved theme would flash', () => {
    expect(themeInitSource).toContain(`'${THEME_STORAGE_KEY}'`);
  });

  it.each(['light', 'dark'])('applies a saved %s theme', (theme) => {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);

    runThemeInit();

    expect(document.documentElement).toHaveAttribute(THEME_ATTRIBUTE, theme);
  });

  it.each([null, 'system', 'nonsense'])('leaves the system theme alone for %s', (saved) => {
    if (saved !== null) {
      window.localStorage.setItem(THEME_STORAGE_KEY, saved);
    }

    runThemeInit();

    expect(document.documentElement).not.toHaveAttribute(THEME_ATTRIBUTE);
  });
});
