import { describe, expect, it } from 'vitest';
import { HISTORY_LIMIT } from '../../constants/calculator';
import { addHistoryEntry, type HistoryEntry } from './history';

function entry(id: number): HistoryEntry {
  return { id, request: { operation: 'add', operands: [String(id), '1'] }, result: String(id + 1) };
}

describe('addHistoryEntry', () => {
  it('puts the newest entry first', () => {
    const history = addHistoryEntry([entry(1)], entry(2));

    expect(history.map((item) => item.id)).toEqual([2, 1]);
  });

  it('keeps every entry below the limit', () => {
    let history: HistoryEntry[] = [];
    for (let id = 1; id <= HISTORY_LIMIT; id += 1) {
      history = addHistoryEntry(history, entry(id));
    }

    expect(history).toHaveLength(HISTORY_LIMIT);
  });

  it('drops the oldest entry past the limit of 20', () => {
    let history: HistoryEntry[] = [];
    for (let id = 1; id <= HISTORY_LIMIT + 1; id += 1) {
      history = addHistoryEntry(history, entry(id));
    }

    expect(HISTORY_LIMIT).toBe(20);
    expect(history).toHaveLength(HISTORY_LIMIT);
    expect(history[0]?.id).toBe(HISTORY_LIMIT + 1);
    expect(history.at(-1)?.id).toBe(2);
  });

  it('does not change the array it was given', () => {
    const original = [entry(1)];

    addHistoryEntry(original, entry(2));

    expect(original).toHaveLength(1);
  });
});
