import { describe, expect, it } from 'vitest';
import { darkTokens, lightTokens, rawTokens, resolveLightDark, tokensSource } from '../test/tokens';

describe('tokens.css', () => {
  it('defines each colour once as light-dark(), so the explicit and system dark themes cannot drift', () => {
    // The old file kept two copies of the dark values. light-dark() has one, and color-scheme
    // picks the side: 'light dark' follows the system, [data-theme] pins one.
    const colours = Object.entries(rawTokens).filter(([name]) => name.startsWith('--color-'));

    expect(colours.length).toBeGreaterThan(0);
    for (const [name, value] of colours) {
      expect(value, name).toMatch(/^light-dark\(/);
    }
    expect(tokensSource.match(/--color-bg:/g)).toHaveLength(1);
  });

  it('follows the system by default and lets an explicit choice win', () => {
    expect(tokensSource).toMatch(/:root\s*\{\s*color-scheme:\s*light dark;/);
    expect(tokensSource).toMatch(/:root\[data-theme=['"]light['"]\]\s*\{\s*color-scheme:\s*light;/);
    expect(tokensSource).toMatch(/:root\[data-theme=['"]dark['"]\]\s*\{\s*color-scheme:\s*dark;/);
  });

  it('resolves nested light-dark() calls to the side of the scheme', () => {
    expect(resolveLightDark('0 1px light-dark(rgb(1 2 3 / 4%), #000)', 'light')).toBe(
      '0 1px rgb(1 2 3 / 4%)',
    );
    expect(resolveLightDark('light-dark(#fff, #000) light-dark(a, b)', 'dark')).toBe('#000 b');
    expect(() => resolveLightDark('light-dark(#fff', 'dark')).toThrow('Unbalanced');
    expect(() => resolveLightDark('light-dark(#fff)', 'dark')).toThrow('two arguments');
  });

  it('gives the two schemes different surfaces and text', () => {
    for (const name of ['--color-bg', '--color-card', '--color-text', '--color-accent']) {
      expect(darkTokens[name], name).not.toBe(lightTokens[name]);
    }
  });

  it('uses the colours from the brief', () => {
    expect(lightTokens).toMatchObject({
      '--color-bg': '#f7f6f3',
      '--color-key': '#ffffff',
      '--color-text': '#111827',
      '--color-border': '#e5e7eb',
      '--color-accent': '#5b3df5',
    });
    expect(darkTokens).toMatchObject({
      '--color-bg': '#0b0f14',
      '--color-key': '#151b23',
      '--color-text': '#e6edf3',
      '--color-accent': '#2dd4a7',
    });
  });

  it('keeps every motion duration under 200ms', () => {
    for (const [name, value] of Object.entries(lightTokens)) {
      if (name.startsWith('--duration-')) {
        expect(Number.parseInt(value, 10), name).toBeLessThan(200);
      }
    }
  });
});
