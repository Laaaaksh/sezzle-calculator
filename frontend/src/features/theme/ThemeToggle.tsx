import type { ThemePreference } from './theme';
import styles from './ThemeToggle.module.css';
import { useTheme } from './useTheme';

const OPTIONS: readonly { value: ThemePreference; label: string }[] = [
  { value: 'system', label: 'System' },
  { value: 'light', label: 'Light' },
  { value: 'dark', label: 'Dark' },
];

/**
 * Three buttons instead of one that cycles: the state is always visible and announced
 * (aria-pressed), and nobody has to press repeatedly to find the option they want.
 */
export function ThemeToggle() {
  const { preference, setPreference } = useTheme();

  return (
    <div className={styles.group} role="group" aria-label="Theme">
      {OPTIONS.map(({ value, label }) => (
        <button
          key={value}
          type="button"
          className={styles.option}
          aria-pressed={preference === value}
          onClick={() => {
            setPreference(value);
          }}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
