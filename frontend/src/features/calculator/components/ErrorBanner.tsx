import { messageForError } from '../../../constants/errorMessages';
import type { CalculatorError } from '../state';
import styles from './ErrorBanner.module.css';

interface ErrorBannerProps {
  error: CalculatorError;
  /** Sends the same request again. Only offered when the error is one a retry can fix. */
  onRetry: () => void;
}

export function ErrorBanner({ error, onRetry }: ErrorBannerProps) {
  return (
    <div className={styles.banner}>
      <p className={styles.message}>{messageForError(error)}</p>
      {error.retryable ? (
        <button type="button" className={styles.retry} onClick={onRetry}>
          Retry
        </button>
      ) : null}
    </div>
  );
}
