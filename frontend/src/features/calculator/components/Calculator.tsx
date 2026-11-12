import { useCalculator, type CalculatorClient } from '../useCalculator';
import { useCalculatorKeyboard } from '../useCalculatorKeyboard';
import { activeOperatorOf } from '../display';
import styles from './Calculator.module.css';
import { Display } from './Display';
import { ErrorBanner } from './ErrorBanner';
import { HistoryDrawer } from './HistoryDrawer';
import { Keypad } from './Keypad';

interface CalculatorProps {
  /** Tests pass a client aimed at their mock server. The app uses the default. */
  client?: CalculatorClient | undefined;
}

export function Calculator({ client }: CalculatorProps) {
  const { state, history, isPending, dispatch, reuse } = useCalculator({ client });
  useCalculatorKeyboard(dispatch);

  return (
    <div className={styles.layout}>
      <div className={styles.card}>
        <Display state={state} isPending={isPending} />
        {state.status === 'error' ? (
          <ErrorBanner
            error={state.error}
            onRetry={() => {
              dispatch({ type: 'retry' });
            }}
          />
        ) : null}
        <Keypad isPending={isPending} activeOperator={activeOperatorOf(state)} onPress={dispatch} />
      </div>
      <HistoryDrawer entries={history} onReuse={reuse} />
    </div>
  );
}
