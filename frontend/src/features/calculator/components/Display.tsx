import type { ReactNode } from 'react';
import { describeDisplay } from '../display';
import type { CalculatorState } from '../state';
import styles from './Display.module.css';

/** Values longer than this get a size hint, so the stylesheet can shrink the text to fit. */
const LONG_VALUE_LENGTH = 12;
const VERY_LONG_VALUE_LENGTH = 24;

interface DisplayProps {
  state: CalculatorState;
  isPending: boolean;
  /** Rendered inside the live region so an error message is announced with the display. */
  children?: ReactNode;
}

function sizeOf(value: string): 'normal' | 'long' | 'very-long' {
  if (value.length > VERY_LONG_VALUE_LENGTH) {
    return 'very-long';
  }
  return value.length > LONG_VALUE_LENGTH ? 'long' : 'normal';
}

export function Display({ state, isPending, children }: DisplayProps) {
  const { expression, value } = describeDisplay(state);
  const valueClass = [styles.value, isPending ? styles.pending : undefined]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={styles.display} role="status" aria-live="polite" aria-busy={isPending}>
      <p className={styles.expression}>{expression}</p>
      <p className={valueClass} data-size={sizeOf(value)}>
        {value}
      </p>
      {children}
    </div>
  );
}
