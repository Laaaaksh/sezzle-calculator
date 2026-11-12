import { fireEvent, render, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CalculatorAction } from './state';
import { useCalculatorKeyboard } from './useCalculatorKeyboard';

function setup(enabled = true) {
  const onAction = vi.fn<(action: CalculatorAction) => void>();
  const hook = renderHook(() => {
    useCalculatorKeyboard(onAction, enabled);
  });
  return { onAction, ...hook };
}

describe('useCalculatorKeyboard', () => {
  it('dispatches the action for a key pressed on the page', () => {
    const { onAction } = setup();

    fireEvent.keyDown(document.body, { key: '7' });

    expect(onAction).toHaveBeenCalledExactlyOnceWith({ type: 'digit', digit: '7' });
  });

  it('prevents the browser default for a handled key', () => {
    setup();

    const notPrevented = fireEvent.keyDown(document.body, { key: '/' });

    expect(notPrevented).toBe(false);
  });

  it('leaves an unmapped key alone', () => {
    const { onAction } = setup();

    const notPrevented = fireEvent.keyDown(document.body, { key: 'a' });

    expect(onAction).not.toHaveBeenCalled();
    expect(notPrevented).toBe(true);
  });

  it.each([
    ['Ctrl', { ctrlKey: true }],
    ['Meta', { metaKey: true }],
    ['Alt', { altKey: true }],
  ])('ignores a key pressed with %s', (_label, modifier) => {
    const { onAction } = setup();

    fireEvent.keyDown(document.body, { key: 'r', ...modifier });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('ignores a key typed inside an input, so it does not fire twice', () => {
    const { onAction } = setup();
    const { getByLabelText } = render(<input aria-label="note" />);

    fireEvent.keyDown(getByLabelText('note'), { key: '5' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('ignores a key typed inside a textarea', () => {
    const { onAction } = setup();
    const { getByLabelText } = render(<textarea aria-label="note" />);

    fireEvent.keyDown(getByLabelText('note'), { key: '5' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('ignores a key typed inside an editable element', () => {
    const { onAction } = setup();
    const { getByTestId } = render(
      <div contentEditable suppressContentEditableWarning data-testid="editor">
        <span data-testid="inner">text</span>
      </div>,
    );

    fireEvent.keyDown(getByTestId('inner'), { key: '5' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('ignores Enter on a focused button, which the button handles itself', () => {
    const { onAction } = setup();
    const { getByRole } = render(<button type="button">key</button>);

    fireEvent.keyDown(getByRole('button'), { key: 'Enter' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('ignores Space on a focused button', () => {
    const { onAction } = setup();
    const { getByRole } = render(<button type="button">key</button>);

    fireEvent.keyDown(getByRole('button'), { key: ' ' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('still handles a digit pressed while a button has focus', () => {
    const { onAction } = setup();
    const { getByRole } = render(<button type="button">key</button>);

    fireEvent.keyDown(getByRole('button'), { key: '3' });

    expect(onAction).toHaveBeenCalledExactlyOnceWith({ type: 'digit', digit: '3' });
  });

  it('ignores a key press another handler already claimed', () => {
    const { onAction } = setup();
    const claim = (event: KeyboardEvent) => {
      event.preventDefault();
    };
    document.body.addEventListener('keydown', claim);

    fireEvent.keyDown(document.body, { key: '5' });
    document.body.removeEventListener('keydown', claim);

    expect(onAction).not.toHaveBeenCalled();
  });

  it('does nothing while disabled', () => {
    const { onAction } = setup(false);

    fireEvent.keyDown(document.body, { key: '5' });

    expect(onAction).not.toHaveBeenCalled();
  });

  it('stops listening after unmount', () => {
    const { onAction, unmount } = setup();

    unmount();
    fireEvent.keyDown(document.body, { key: '5' });

    expect(onAction).not.toHaveBeenCalled();
  });
});
