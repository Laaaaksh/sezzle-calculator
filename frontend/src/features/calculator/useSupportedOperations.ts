import { useEffect, useState } from 'react';
import { getOperations, type ApiClient, type OperationType } from '../../api';

/** The part of the API client this hook needs. Tests pass their own. */
export type OperationsClient = Pick<ApiClient, 'getOperations'>;

const DEFAULT_CLIENT: OperationsClient = { getOperations };

/**
 * The operations the backend reports (GET /api/v1/operations), fetched once on mount.
 * Returns null while loading and when the fetch fails or the list is empty. Null means "unknown",
 * and the keypad then leaves every key enabled: the keypad renders at once from the static key
 * list, and the server still validates every request, so a failed lookup must never lock the
 * calculator or show an error.
 */
export function useSupportedOperations(
  client: OperationsClient = DEFAULT_CLIENT,
): ReadonlySet<OperationType> | null {
  const [supported, setSupported] = useState<ReadonlySet<OperationType> | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    void client.getOperations({ signal: controller.signal }).then(
      (operations) => {
        // An empty list would disable every key, so it is treated like a failed lookup.
        if (!controller.signal.aborted && operations.length > 0) {
          setSupported(new Set(operations.map((operation) => operation.name)));
        }
      },
      () => {
        // Deliberately ignored: see above. The keypad keeps working without the list.
      },
    );
    return () => {
      controller.abort();
    };
  }, [client]);

  return supported;
}
