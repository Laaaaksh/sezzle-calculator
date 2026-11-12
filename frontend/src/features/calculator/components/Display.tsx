import type { CSSProperties, Ref } from 'react';
import { describeDisplay } from '../display';
import type { CalculatorState } from '../state';
import styles from './Display.module.css';

interface DisplayProps {
  state: CalculatorState;
  isPending: boolean;
  /** Lets the caller move keyboard focus here, for example after Retry removes the button. */
  ref?: Ref<HTMLDivElement>;
}

export function Display({ state, isPending, ref }: DisplayProps) {
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

  // tabIndex -1: not in the tab order, but a script can focus it (see Calculator's retry).
  // Only the value is a live region. Announcing the expression too would read both on every key.
  return (
    <div ref={ref} className={styles.display} role="group" aria-label="Display" tabIndex={-1}>
      <p className={styles.expression}>{expression}</p>
      {/* The live region stays mounted. Only the paragraph inside is replaced, so the change is read. */}
      <div role="status" aria-live="polite" aria-busy={isPending}>
        <p key={animationKey} className={valueClass} style={valueStyle}>
          {value}
        </p>
      </div>
    </div>
  );
}
