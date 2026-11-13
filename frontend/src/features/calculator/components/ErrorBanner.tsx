import { messageForError } from '../../../constants/errorMessages';
import type { CalculatorError } from '../state';
import styles from './ErrorBanner.module.css';

interface ErrorBannerProps {
  error: CalculatorError;
  onRetry: () => void;
}

/**
 * The message is an alert, which assistive tech reads out at once. A polite status region could
 * queue it behind other updates, and an error should not wait.
 */
export function ErrorBanner({ error, onRetry }: ErrorBannerProps) {
  return (
    <div className={styles.banner}>
      <p className={styles.message} role="alert">
        {messageForError(error)}
      </p>
      {error.retryable ? (
        <button type="button" className={styles.retry} onClick={onRetry}>
          Retry
        </button>
      ) : null}
    </div>
  );
}
