import { useEffect, useEffectEvent, useState } from 'react';
import { useApiClient, type OperationLimits } from '../../api';

export interface OperationsCatalog {
  /** Names of the operations the server offers. Null = unknown (loading or the fetch failed). */
  supported: ReadonlySet<string> | null;
  /** The server's input limits. Undefined when unknown or when an older server omits them. */
  limits: OperationLimits | undefined;
}

const UNKNOWN_CATALOG: OperationsCatalog = { supported: null, limits: undefined };

/**
 * What the backend reports (GET /api/v1/operations), fetched once on mount.
 * Everything stays "unknown" while loading and when the fetch fails or the list is empty. Unknown
 * means the keypad leaves every key enabled and uses its built-in limits: the keypad renders at
 * once from the static key list, and the server still validates every request, so a failed
 * lookup must never lock the calculator or show an error.
 */
export function useOperationsCatalog(): OperationsCatalog {
  const client = useApiClient();
  const [catalog, setCatalog] = useState<OperationsCatalog>(UNKNOWN_CATALOG);
  // An effect event reads the latest client without being a dependency, so an inline client
  // object (a new identity every render) cannot make the effect fetch again.
  const fetchOperations = useEffectEvent((signal: AbortSignal) => client.getOperations({ signal }));

  useEffect(() => {
    const controller = new AbortController();
    void fetchOperations(controller.signal).then(
      ({ operations, limits }) => {
        // An empty list would disable every key, so it is treated like a failed lookup.
        if (!controller.signal.aborted && operations.length > 0) {
          setCatalog({ supported: new Set(operations.map((operation) => operation.name)), limits });
        }
      },
      () => {
        // Deliberately ignored: see above. The keypad keeps working without the list.
      },
    );
    return () => {
      controller.abort();
    };
  }, []);

  return catalog;
}
