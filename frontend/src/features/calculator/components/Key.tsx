import { useId } from 'react';
import type { KeyDefinition } from '../keys';
import type { CalculatorAction } from '../state';
import styles from './Key.module.css';

interface KeyProps {
  definition: KeyDefinition;
  disabled?: boolean;
  /** The server does not offer this operation. Stays focusable so the reason can be read. */
  unavailable?: boolean;
  /** This operator is waiting for its second operand. Shown with a ring, not colour alone. */
  active?: boolean;
  onPress: (action: CalculatorAction) => void;
}

const UNAVAILABLE_REASON = 'Not available on this server';

export function Key({
  definition,
  disabled = false,
  unavailable = false,
  active = false,
  onPress,
}: KeyProps) {
  const { label, ariaLabel, action, kind, wide = false } = definition;
  const reasonId = useId();
  const className = [
    styles.key,
    styles[kind],
    wide ? styles.wide : undefined,
    active ? styles.active : undefined,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <>
      <button
        type="button"
        className={className}
        aria-label={ariaLabel}
        // Only operators can be "on", so only they announce a pressed state.
        aria-pressed={kind === 'operator' ? active : undefined}
        disabled={disabled}
        // aria-disabled, not disabled: the key stays in the tab order, so a keyboard or screen
        // reader user can land on it and hear the reason from the title.
        aria-disabled={unavailable ? true : undefined}
        // The title alone is not announced reliably, so the same reason is also a hidden element
        // that describes the button. Both stay: the title serves pointer users.
        aria-describedby={unavailable ? reasonId : undefined}
        title={unavailable ? UNAVAILABLE_REASON : undefined}
        // Keeps focus off the pad after a pointer press. Otherwise Enter would press this key again
        // instead of acting as equals. Keyboard users still reach every key with Tab.
        onMouseDown={(event) => {
          event.preventDefault();
        }}
        onClick={() => {
          if (!unavailable) {
            onPress(action);
          }
        }}
      >
        {label}
      </button>
      {unavailable ? (
        <span id={reasonId} className="visually-hidden">
          {UNAVAILABLE_REASON}
        </span>
      ) : null}
    </>
  );
}
