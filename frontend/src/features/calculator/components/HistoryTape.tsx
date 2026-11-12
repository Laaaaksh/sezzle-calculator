import { formatExpression } from '../expression';
import type { HistoryEntry } from '../history';
import styles from './HistoryTape.module.css';

interface HistoryTapeProps {
  /** Newest first. */
  entries: readonly HistoryEntry[];
  onReuse: (entry: HistoryEntry) => void;
}

export function HistoryTape({ entries, onReuse }: HistoryTapeProps) {
  return (
    <section aria-labelledby="history-heading">
      <h2 id="history-heading" className={styles.heading}>
        History
      </h2>
      {entries.length === 0 ? (
        <p className={styles.empty}>No calculations yet.</p>
      ) : (
        <ol className={styles.list}>
          {entries.map((entry) => (
            <li key={entry.id}>
              <button
                type="button"
                className={styles.entry}
                onMouseDown={(event) => {
                  event.preventDefault();
                }}
                onClick={() => {
                  onReuse(entry);
                }}
              >
                <span className={styles.hidden}>Reuse </span>
                <span className={styles.expression}>{formatExpression(entry.request)}</span>
                {' = '}
                <span className={styles.result}>{entry.result}</span>
              </button>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
