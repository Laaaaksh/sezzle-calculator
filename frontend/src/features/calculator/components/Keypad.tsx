import type { BinaryOperation } from '../../../constants/calculator';
import { KEYS } from '../keys';
import type { CalculatorAction } from '../state';
import { Key } from './Key';
import styles from './Keypad.module.css';

interface KeypadProps {
  /** While true, equals is disabled so a second request cannot be submitted. */
  isPending: boolean;
  /** The operator waiting for its second operand, if any. */
  activeOperator?: BinaryOperation | null | undefined;
  onPress: (action: CalculatorAction) => void;
}

export function Keypad({ isPending, activeOperator = null, onPress }: KeypadProps) {
  return (
    <div className={styles.keypad} role="group" aria-label="Calculator keys">
      {KEYS.map((definition) => (
        <Key
          key={definition.id}
          definition={definition}
          disabled={isPending && definition.action.type === 'equals'}
          active={
            definition.action.type === 'operator' && definition.action.operator === activeOperator
          }
          onPress={onPress}
        />
      ))}
    </div>
  );
}
