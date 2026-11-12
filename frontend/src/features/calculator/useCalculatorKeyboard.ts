import { useEffect } from 'react';
import { actionForKey } from './keyboard';
import type { CalculatorAction } from './state';

const TEXT_FIELD_SELECTOR =
  'input, textarea, select, [contenteditable=""], [contenteditable="true"]';
/** Enter and Space activate a focused button on their own, so the shortcut must stay out of it. */
const BUTTON_ACTIVATION_KEYS = ['Enter', ' '];

function isTextField(target: EventTarget | null): boolean {
  return target instanceof Element && target.closest(TEXT_FIELD_SELECTOR) !== null;
}

function isActivatingButton(event: KeyboardEvent): boolean {
  return (
    event.target instanceof Element &&
    event.target.closest('button') !== null &&
    BUTTON_ACTIVATION_KEYS.includes(event.key)
  );
}

/**
 * Turns key presses on the page into calculator actions.
 * It stays out of the way whenever the browser already handles the key, so nothing fires twice:
 * typing in a text field, Enter or Space on a focused button, and shortcuts like Ctrl+R.
 */
export function useCalculatorKeyboard(
  onAction: (action: CalculatorAction) => void,
  enabled = true,
): void {
  useEffect(() => {
    if (!enabled) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }
      if (isTextField(event.target) || isActivatingButton(event)) {
        return;
      }
      const action = actionForKey(event.key);
      if (action === null) {
        return;
      }
      // Stops the browser's own use of the key, such as quick find on "/".
      event.preventDefault();
      onAction(action);
    };
    window.addEventListener('keydown', onKeyDown);
    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, [onAction, enabled]);
}
