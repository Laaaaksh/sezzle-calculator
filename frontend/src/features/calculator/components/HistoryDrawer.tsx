import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react';
import { WIDE_LAYOUT_QUERY } from '../../../constants/layout';
import type { HistoryEntry } from '../history';
import styles from './HistoryDrawer.module.css';
import { HistoryTape } from './HistoryTape';

interface HistoryDrawerProps {
  /** Newest first. */
  entries: readonly HistoryEntry[];
  onReuse: (entry: HistoryEntry) => void;
}

/**
 * Wide screens show the tape beside the keypad and hide the button. Narrow screens hide the tape
 * behind it. The stylesheet decides which, so the two layouts never fight over state.
 */
export function HistoryDrawer({ entries, onReuse }: HistoryDrawerProps) {
  const [open, setOpen] = useState(false);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') {
      return undefined;
    }
    const wide = window.matchMedia(WIDE_LAYOUT_QUERY);
    const onChange = (event: MediaQueryListEvent) => {
      if (event.matches) {
        setOpen(false);
      }
    };
    wide.addEventListener('change', onChange);
    return () => {
      wide.removeEventListener('change', onChange);
    };
  }, []);

  const close = () => {
    setOpen(false);
    // Focus would otherwise vanish with the hidden panel and drop keyboard users at the page top.
    toggleRef.current?.focus();
  };

  const onKeyDown = (event: KeyboardEvent) => {
    if (open && event.key === 'Escape') {
      // Marks the key as handled so the calculator's own Escape shortcut (clear) stays quiet.
      event.preventDefault();
      close();
    }
  };

  return (
    // The handler only listens for Escape bubbling up from the button and panel; nothing here is
    // itself interactive.
    // eslint-disable-next-line jsx-a11y/no-static-element-interactions
    <div className={styles.drawer} onKeyDown={onKeyDown}>
      <button
        ref={toggleRef}
        type="button"
        className={styles.toggle}
        aria-expanded={open}
        aria-controls={panelId}
        onClick={() => {
          setOpen((current) => !current);
        }}
      >
        History
        <span className={styles.chevron} aria-hidden="true" />
      </button>
      <div id={panelId} className={styles.panel} data-open={open}>
        <HistoryTape
          entries={entries}
          headingClassName={styles.panelHeading}
          onReuse={(entry) => {
            onReuse(entry);
            if (open) {
              close();
            }
          }}
        />
      </div>
    </div>
  );
}
