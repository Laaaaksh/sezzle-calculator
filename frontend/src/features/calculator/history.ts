import { HISTORY_LIMIT } from '../../constants/calculator';
import type { CalculationRequest } from './state';

export interface HistoryEntry {
  /** The id of the request that produced it. Unique, so it works as a list key. */
  id: number;
  request: CalculationRequest;
  /** The server's string, never parsed into a JS number. */
  result: string;
}

/** Puts the new entry first and drops the oldest ones past the limit. Returns a new array. */
export function addHistoryEntry(
  history: readonly HistoryEntry[],
  entry: HistoryEntry,
  limit = HISTORY_LIMIT,
): HistoryEntry[] {
  return [entry, ...history].slice(0, limit);
}
