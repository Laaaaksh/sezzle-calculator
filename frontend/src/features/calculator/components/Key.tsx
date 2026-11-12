import type { KeyDefinition } from '../keys';
import type { CalculatorAction } from '../state';
import styles from './Key.module.css';

interface KeyProps {
  definition: KeyDefinition;
  disabled?: boolean;
  /** This operator is waiting for its second operand. Shown with a ring, not colour alone. */
  active?: boolean;
  onPress: (action: CalculatorAction) => void;
}

export function Key({ definition, disabled = false, active = false, onPress }: KeyProps) {
  const { label, ariaLabel, action, kind, wide = false } = definition;
  const className = [
    styles.key,
    styles[kind],
    wide ? styles.wide : undefined,
    active ? styles.active : undefined,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <button
      type="button"
      className={className}
      aria-label={ariaLabel}
      // Only operators can be "on", so only they announce a pressed state.
      aria-pressed={kind === 'operator' ? active : undefined}
      disabled={disabled}
      // Keeps focus off the pad after a pointer press. Otherwise Enter would press this key again
      // instead of acting as equals. Keyboard users still reach every key with Tab.
      onMouseDown={(event) => {
        event.preventDefault();
      }}
      onClick={() => {
        onPress(action);
      }}
    >
      {label}
    </button>
  );
}
