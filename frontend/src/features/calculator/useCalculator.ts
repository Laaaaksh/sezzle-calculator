import {
  useCallback,
  useEffect,
  useEffectEvent,
  useMemo,
  useReducer,
  useState,
  type Dispatch,
} from 'react';
import { CanceledError, useApiClient, type CalculateRequest } from '../../api';
import { MAX_OPERAND_LENGTH } from '../../constants/calculator';
import { addHistoryEntry, type HistoryEntry } from './history';
import {
  createCalculatorReducer,
  initialState,
  toCalculatorError,
  type CalculatorAction,
  type CalculatorState,
} from './state';

export interface UseCalculatorOptions {
  /**
   * Longest operand the server accepts (from its limits). Defaults to MAX_OPERAND_LENGTH, which
   * is also what applies when the server does not report a limit.
   */
  maxOperandLength?: number | undefined;
}

export interface UseCalculatorResult {
  state: CalculatorState;
  /** Newest first, capped at HISTORY_LIMIT. Lives in memory only. */
  history: readonly HistoryEntry[];
  /** True while a request is in flight. Equals and the other keys are ignored until it ends. */
  isPending: boolean;
  dispatch: Dispatch<CalculatorAction>;
  /** Puts a history entry's result back on the display. */
  reuse: (entry: HistoryEntry) => void;
}

/**
 * Wires the reducer to the API: it sends what the reducer asks for and reports the outcome.
 * The client comes from ApiClientContext.
 */
export function useCalculator({
  maxOperandLength = MAX_OPERAND_LENGTH,
}: UseCalculatorOptions = {}): UseCalculatorResult {
  const client = useApiClient();
  const reducer = useMemo(() => createCalculatorReducer({ maxOperandLength }), [maxOperandLength]);
  const [state, dispatch] = useReducer(reducer, initialState);
  const [history, setHistory] = useState<readonly HistoryEntry[]>([]);
  const pending = state.status === 'waiting' ? state.pending : null;
  // An effect event reads the latest client without being a dependency: a client with a new
  // identity on every render must not abort and re-send the request that is in flight.
  const sendCalculation = useEffectEvent((request: CalculateRequest, signal: AbortSignal) =>
    client.calculate(request, { signal }),
  );

  useEffect(() => {
    if (pending === null) {
      return;
    }
    // The cleanup below aborts this request when the state leaves "waiting" (clear, retry,
    // unmount). The reducer also checks the request id, so a late answer is dropped twice over.
    const controller = new AbortController();
    const request = { operation: pending.operation, operands: [...pending.operands] };

    void sendCalculation(request, controller.signal).then(
      ({ result }) => {
        if (controller.signal.aborted) {
          return;
        }
        dispatch({ type: 'result_received', requestId: pending.id, result });
        setHistory((entries) => addHistoryEntry(entries, { id: pending.id, request, result }));
      },
      (error: unknown) => {
        // A canceled request is the user's own doing (clear), not a failure to show.
        if (controller.signal.aborted || error instanceof CanceledError) {
          return;
        }
        dispatch({
          type: 'request_failed',
          requestId: pending.id,
          error: toCalculatorError(error),
        });
      },
    );

    return () => {
      controller.abort();
    };
  }, [pending]);

  const reuse = useCallback((entry: HistoryEntry) => {
    dispatch({ type: 'reuse', result: entry.result });
  }, []);

  return { state, history, isPending: pending !== null, dispatch, reuse };
}
