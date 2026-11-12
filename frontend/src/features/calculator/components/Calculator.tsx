import { useCallback, useRef } from 'react';
import { isActionAvailable } from '../keys';
import type { CalculatorAction } from '../state';
import { useCalculator, type CalculatorClient } from '../useCalculator';
import { useCalculatorKeyboard } from '../useCalculatorKeyboard';
import { useSupportedOperations, type OperationsClient } from '../useSupportedOperations';
import { activeOperatorOf } from '../display';
import styles from './Calculator.module.css';
import { Display } from './Display';
import { ErrorBanner } from './ErrorBanner';
import { HistoryDrawer } from './HistoryDrawer';
import { Keypad } from './Keypad';

interface CalculatorProps {
  /** Tests pass a client aimed at their mock server. The app uses the default. */
  client?: (CalculatorClient & OperationsClient) | undefined;
}

export function Calculator({ client }: CalculatorProps) {
  const { state, history, isPending, dispatch, reuse } = useCalculator({ client });
  const supportedOperations = useSupportedOperations(client);
  const displayRef = useRef<HTMLDivElement>(null);

  // One gate for the keypad and the keyboard shortcuts, so neither can start an operation the
  // server does not offer.
  const press = useCallback(
    (action: CalculatorAction) => {
      if (isActionAvailable(action, supportedOperations)) {
        dispatch(action);
      }
    },
    [dispatch, supportedOperations],
  );
  useCalculatorKeyboard(press);

  // A refused key (result too long to reuse) is shown like an error, but keeps the result.
  const banner =
    state.status === 'error'
      ? state.error
      : state.status === 'showing_result'
        ? state.notice
        : undefined;

  return (
    <div className={styles.layout}>
      <div className={styles.card}>
        <Display ref={displayRef} state={state} isPending={isPending} />
        {banner === undefined ? null : (
          <ErrorBanner
            error={banner}
            onRetry={() => {
              dispatch({ type: 'retry' });
              // Retry unmounts its own button, which would drop focus to the page. The display is
              // the stable place where the retried calculation's result appears. Equals is no
              // option: it is disabled while the request is pending, so it cannot take focus.
              displayRef.current?.focus();
            }}
          />
        )}
        <Keypad
          isPending={isPending}
          activeOperator={activeOperatorOf(state)}
          supportedOperations={supportedOperations}
          onPress={press}
        />
      </div>
      <HistoryDrawer entries={history} onReuse={reuse} />
    </div>
  );
}
