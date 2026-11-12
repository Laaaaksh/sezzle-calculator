import { useCallback, useEffect, useState } from 'react';
import { applyTheme, readStoredTheme, storeTheme, type ThemePreference } from './theme';

export interface UseThemeResult {
  preference: ThemePreference;
  setPreference: (preference: ThemePreference) => void;
}

export function useTheme(): UseThemeResult {
  const [preference, setPreferenceState] = useState<ThemePreference>(readStoredTheme);

  useEffect(() => {
    applyTheme(preference);
  }, [preference]);

  const setPreference = useCallback((next: ThemePreference) => {
    setPreferenceState(next);
    storeTheme(next);
  }, []);

  return { preference, setPreference };
}
