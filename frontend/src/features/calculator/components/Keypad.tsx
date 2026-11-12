import { KEYS } from '../keys';
import type { CalculatorAction } from '../state';
import { Key } from './Key';
import styles from './Keypad.module.css';

interface KeypadProps {
  /** While true, equals is disabled so a second request cannot be submitted. */
  isPending: boolean;
  onPress: (action: CalculatorAction) => void;
}

export function Keypad({ isPending, onPress }: KeypadProps) {
  return (
    <div className={styles.keypad} role="group" aria-label="Calculator keys">
      {KEYS.map((definition) => (
        <Key
          key={definition.id}
          definition={definition}
          disabled={isPending && definition.action.type === 'equals'}
          onPress={onPress}
        />
      ))}
    </div>
  );
}
