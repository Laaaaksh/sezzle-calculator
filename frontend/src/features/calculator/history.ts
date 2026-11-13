import { HISTORY_LIMIT } from '../../constants/calculator';
import type { CalculationRequest } from './state';

export interface HistoryEntry {
  /** The id of the request that produced it. Unique, so it works as a list key. */
  id: number;
  request: CalculationRequest;
  /** The server's string, never parsed into a JS number. */
  result: string;
}

export function addHistoryEntry(
  history: readonly HistoryEntry[],
  entry: HistoryEntry,
  limit = HISTORY_LIMIT,
): HistoryEntry[] {
  return [entry, ...history].slice(0, limit);
}
