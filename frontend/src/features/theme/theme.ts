export const THEME_PREFERENCES = ['system', 'light', 'dark'] as const;
export type ThemePreference = (typeof THEME_PREFERENCES)[number];

export const DEFAULT_THEME_PREFERENCE: ThemePreference = 'system';

/** Must equal THEME_KEY in public/theme-init.js, which applies the saved theme before first paint. */
export const THEME_STORAGE_KEY = 'calculator-theme';
export const THEME_ATTRIBUTE = 'data-theme';

export function isThemePreference(value: unknown): value is ThemePreference {
  return THEME_PREFERENCES.some((preference) => preference === value);
}

/** The saved choice, or the default when nothing valid is saved or storage is blocked. */
export function readStoredTheme(): ThemePreference {
  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
    return isThemePreference(stored) ? stored : DEFAULT_THEME_PREFERENCE;
  } catch {
    // Private mode or blocked storage can throw on access. Following the system is a fine default.
    return DEFAULT_THEME_PREFERENCE;
  }
}

export function storeTheme(preference: ThemePreference): void {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, preference);
  } catch {
    // The choice still applies for this visit; it just will not survive a reload.
  }
}

/**
 * 'system' removes the attribute so the stylesheet's prefers-color-scheme rule decides.
 * 'light' and 'dark' set it, which wins over the system.
 */
export function applyTheme(
  preference: ThemePreference,
  root: HTMLElement = document.documentElement,
): void {
  if (preference === 'system') {
    root.removeAttribute(THEME_ATTRIBUTE);
  } else {
    root.setAttribute(THEME_ATTRIBUTE, preference);
  }
}
