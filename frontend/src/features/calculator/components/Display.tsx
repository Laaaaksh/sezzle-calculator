import type { CSSProperties } from 'react';
import { describeDisplay } from '../display';
import type { CalculatorState } from '../state';
import styles from './Display.module.css';

interface DisplayProps {
  state: CalculatorState;
  isPending: boolean;
}

export function Display({ state, isPending }: DisplayProps) {
  const { expression, value } = describeDisplay(state);
  const valueClass = [
    styles.value,
    isPending ? styles.pending : undefined,
    state.status === 'error' ? styles.error : undefined,
  ]
    .filter(Boolean)
    .join(' ');
  // A new key restarts the entrance animation when a result arrives, but not on every keystroke.
  const animationKey = state.status === 'showing_result' ? `result-${value}` : 'input';
  // The stylesheet sizes the text from this, so long numbers shrink to fit without any script.
  const valueStyle = { '--value-length': Math.max(value.length, 1) } as CSSProperties;

  return (
    <div className={styles.display} role="status" aria-live="polite" aria-busy={isPending}>
      <p className={styles.expression}>{expression}</p>
      <p key={animationKey} className={valueClass} style={valueStyle}>
        {value}
      </p>
    </div>
  );
}
