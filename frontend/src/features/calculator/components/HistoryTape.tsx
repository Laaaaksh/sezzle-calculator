import { useId } from 'react';
import { formatExpression } from '../expression';
import type { HistoryEntry } from '../history';
import styles from './HistoryTape.module.css';

interface HistoryTapeProps {
  /** Newest first. */
  entries: readonly HistoryEntry[];
  onReuse: (entry: HistoryEntry) => void;
  /** Extra class for the heading, so a parent can restyle it without a selector into this component. */
  headingClassName?: string | undefined;
}

export function HistoryTape({ entries, onReuse, headingClassName }: HistoryTapeProps) {
  const headingId = useId();
  return (
    <section aria-labelledby={headingId}>
      <h2 id={headingId} className={[styles.heading, headingClassName].filter(Boolean).join(' ')}>
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
                <span className="visually-hidden">Reuse </span>
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
