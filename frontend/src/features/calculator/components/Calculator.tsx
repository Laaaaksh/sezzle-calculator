import { useCalculator, type CalculatorClient } from '../useCalculator';
import { useCalculatorKeyboard } from '../useCalculatorKeyboard';
import styles from './Calculator.module.css';
import { Display } from './Display';
import { ErrorBanner } from './ErrorBanner';
import { HistoryTape } from './HistoryTape';
import { Keypad } from './Keypad';

interface CalculatorProps {
  /** Tests pass a client aimed at their mock server. The app uses the default. */
  client?: CalculatorClient | undefined;
}

export function Calculator({ client }: CalculatorProps) {
  const { state, history, isPending, dispatch, reuse } = useCalculator({ client });
  useCalculatorKeyboard(dispatch);

  return (
    <div className={styles.calculator}>
      <Display state={state} isPending={isPending}>
        {state.status === 'error' ? (
          <ErrorBanner
            error={state.error}
            onRetry={() => {
              dispatch({ type: 'retry' });
            }}
          />
        ) : null}
      </Display>
      <Keypad isPending={isPending} onPress={dispatch} />
      <HistoryTape entries={history} onReuse={reuse} />
    </div>
  );
}
