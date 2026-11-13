import { describe, expect, it } from 'vitest';
import { declarationsOf, rawTokens, resolveLightDark, ruleFor } from '../test/tokens';

function colorScheme(selector: string): string[] {
  const rule = ruleFor(selector);
  const values: string[] = [];
  rule.walkDecls('color-scheme', (decl) => {
    values.push(decl.value.trim());
  });
  return values;
}

describe('tokens.css', () => {
  it('defines each colour once as light-dark(), so the explicit and system dark themes cannot drift', () => {
    // The old file kept two copies of the dark values. light-dark() has one, and color-scheme
    // picks the side: 'light dark' follows the system, [data-theme] pins one.
    const colours = Object.entries(rawTokens).filter(([name]) => name.startsWith('--color-'));

    expect(colours.length).toBeGreaterThan(0);
    for (const [name, value] of colours) {
      expect(value, name).toMatch(/^light-dark\(/);
    }
    expect(declarationsOf('--color-bg')).toHaveLength(1);
  });

  it('follows the system by default and lets an explicit choice win', () => {
    expect(colorScheme(':root')).toEqual(['light dark']);
    expect(colorScheme(":root[data-theme='light']")).toEqual(['light']);
    expect(colorScheme(":root[data-theme='dark']")).toEqual(['dark']);
  });

  it('resolves nested light-dark() calls to the side of the scheme', () => {
    expect(resolveLightDark('0 1px light-dark(rgb(1 2 3 / 4%), #000)', 'light')).toBe(
      '0 1px rgb(1 2 3 / 4%)',
    );
    expect(resolveLightDark('light-dark(#fff, #000) light-dark(a, b)', 'dark')).toBe('#000 b');
    expect(() => resolveLightDark('light-dark(#fff', 'dark')).toThrow('Unbalanced');
    expect(() => resolveLightDark('light-dark(#fff)', 'dark')).toThrow('two arguments');
  });
});
