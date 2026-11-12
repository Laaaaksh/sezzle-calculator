import { describe, expect, it } from 'vitest';
import { contrastRatio } from '../test/contrast';
import { darkTokens, lightTokens, type TokenMap } from '../test/tokens';

const TEXT_MINIMUM = 4.5;
const UI_MINIMUM = 3;

/** [foreground token, background token, what it is]. Text first, then non-text parts. */
const TEXT_PAIRS: readonly [string, string, string][] = [
  ['--color-text', '--color-bg', 'body text on the page'],
  ['--color-text', '--color-card', 'text on the card'],
  ['--color-text', '--color-well', 'result on the display'],
  ['--color-text', '--color-key', 'digit key label'],
  ['--color-text', '--color-key-hover', 'digit key label on hover'],
  ['--color-text', '--color-key-fn', 'function key label'],
  ['--color-text', '--color-key-fn-hover', 'function key label on hover'],
  ['--color-text-muted', '--color-well', 'expression line on the display'],
  ['--color-text-muted', '--color-card', 'muted text on the card'],
  ['--color-text-muted', '--color-key-hover', 'history expression on hover'],
  ['--color-accent', '--color-key-op', 'operator key label'],
  ['--color-accent', '--color-key-op-hover', 'operator key label on hover'],
  ['--color-on-accent', '--color-accent', 'equals and active operator label'],
  ['--color-on-accent', '--color-accent-hover', 'equals label on hover'],
  ['--color-error-text', '--color-error-bg', 'error message'],
  ['--color-error-text', '--color-well', 'Error on the display'],
  ['--color-error-text', '--color-card', 'Retry button label'],
  ['--color-bg', '--color-text', 'selected theme option'],
  ['--color-text-muted', '--color-card', 'unselected theme option'],
];

/** Parts of the interface that are not text: focus ring and the active operator ring. */
const UI_PAIRS: readonly [string, string, string][] = [
  ['--color-focus', '--color-bg', 'focus ring on the page'],
  ['--color-focus', '--color-card', 'focus ring on the card'],
  ['--color-focus', '--color-key', 'focus ring on a key'],
  ['--color-focus', '--color-key-hover', 'focus ring on a hovered history entry'],
  ['--color-accent', '--color-card', 'active operator ring on the card'],
  ['--color-accent', '--color-key-op', 'active operator fill next to a tinted key'],
  ['--color-accent', '--color-key', 'equals fill next to a digit key'],
];

function ratio(tokens: TokenMap, foreground: string, background: string): number {
  const fg = tokens[foreground];
  const bg = tokens[background];
  if (fg === undefined || bg === undefined) {
    throw new Error(`Missing token: ${foreground} or ${background}`);
  }
  return contrastRatio(fg, bg);
}

describe('contrastRatio', () => {
  it('is 21 for black on white and 1 for identical colours', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 5);
    expect(contrastRatio('#5b3df5', '#5b3df5')).toBeCloseTo(1, 5);
  });

  it('matches a known reference value (#767676 on white is 4.54)', () => {
    expect(contrastRatio('#767676', '#ffffff')).toBeCloseTo(4.54, 2);
  });

  it('does not care which colour is passed first', () => {
    expect(contrastRatio('#111827', '#f7f6f3')).toBe(contrastRatio('#f7f6f3', '#111827'));
  });

  it('rejects anything that is not a six digit hex colour', () => {
    expect(() => contrastRatio('red', '#ffffff')).toThrow(/#rrggbb/);
  });
});

describe.each([
  ['light', lightTokens],
  ['dark', darkTokens],
] as const)('%s theme contrast', (_theme, tokens) => {
  it.each(TEXT_PAIRS)('text %s on %s (%s) is at least 4.5:1', (foreground, background) => {
    expect(ratio(tokens, foreground, background)).toBeGreaterThanOrEqual(TEXT_MINIMUM);
  });

  it.each(UI_PAIRS)('%s against %s (%s) is at least 3:1', (foreground, background) => {
    expect(ratio(tokens, foreground, background)).toBeGreaterThanOrEqual(UI_MINIMUM);
  });
});
