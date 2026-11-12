// Runs as a blocking script in <head>, before the first paint, so a saved theme never flashes the
// wrong colours. It is a separate file, not inline, so a strict Content-Security-Policy
// (script-src 'self') still allows it. Plain JS because it must run before the app bundle loads.
// THEME_KEY must equal THEME_STORAGE_KEY in src/features/theme/theme.ts; themeInit.test.ts checks it.
(function () {
  const THEME_KEY = 'calculator-theme';
  try {
    const saved = window.localStorage.getItem(THEME_KEY);
    if (saved === 'light' || saved === 'dark') {
      document.documentElement.setAttribute('data-theme', saved);
    }
  } catch {
    // Storage blocked (private mode, strict settings): follow the system theme.
  }
})();
