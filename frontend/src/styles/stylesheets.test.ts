import { describe, expect, it } from 'vitest';
import { lightTokens } from '../test/tokens';

/** Every CSS module in the app, as text. Written so a new component is checked without edits. */
const modules = import.meta.glob<string>('../**/*.module.css', {
  query: '?raw',
  import: 'default',
  eager: true,
});
const entries = Object.entries(modules);

interface Rule {
  selector: string;
  body: string;
}

/** Top-level and nested `selector { body }` pairs, with comments removed. */
function rulesOf(css: string): Rule[] {
  const stripped = css.replace(/\/\*[\s\S]*?\*\//g, '');
  return [...stripped.matchAll(/([^{}]+)\{([^{}]*)\}/g)].map((match) => ({
    selector: (match[1] ?? '').trim(),
    body: match[2] ?? '',
  }));
}

/** Removes `@media (...) { ... }` blocks whose query contains `needle`, counting braces. */
function withoutMedia(css: string, needle: string): string {
  let result = '';
  let index = 0;
  while (index < css.length) {
    const start = css.indexOf('@media', index);
    if (start === -1) {
      result += css.slice(index);
      break;
    }
    const open = css.indexOf('{', start);
    const query = css.slice(start, open);
    let depth = 1;
    let end = open + 1;
    while (depth > 0 && end < css.length) {
      depth += css[end] === '{' ? 1 : css[end] === '}' ? -1 : 0;
      end += 1;
    }
    result += css.slice(index, start);
    if (!query.includes(needle)) {
      result += css.slice(start, end);
    }
    index = end;
  }
  return result;
}

describe('the CSS modules were found', () => {
  it('includes the calculator keys', () => {
    expect(entries.some(([path]) => path.endsWith('Key.module.css'))).toBe(true);
  });
});

describe.each(entries)('%s', (_path, css) => {
  it('never removes the focus outline', () => {
    expect(css).not.toMatch(/outline(-style)?\s*:\s*(none|0)\b/);
  });

  it('gives every clickable thing a visible focus ring of at least 2px', () => {
    const clickable = /cursor:\s*pointer/.test(css);
    if (!clickable) {
      return;
    }
    const focusRules = rulesOf(css).filter((rule) => rule.selector.includes(':focus-visible'));
    expect(focusRules.length).toBeGreaterThan(0);
    for (const rule of focusRules) {
      const width = /outline:\s*(\d+)px\s+solid/.exec(rule.body)?.[1];
      // A rule that only moves the offset (the active operator) leans on the base ring.
      if (width !== undefined) {
        expect(Number(width), rule.selector).toBeGreaterThanOrEqual(2);
      }
    }
  });

  it('makes every clickable rule at least 44px tall', () => {
    const clickable = rulesOf(css).filter((rule) => /cursor:\s*pointer/.test(rule.body));
    for (const rule of clickable) {
      expect(rule.body, rule.selector).toMatch(/min-height:\s*var\(--size-(touch|key)\)/);
    }
  });

  it('keeps animation and transitions inside the no-preference reduced motion query', () => {
    const outside = withoutMedia(css, 'prefers-reduced-motion: no-preference');
    expect(outside).not.toMatch(/\b(animation|transition)\s*:/);
  });

  it('takes durations from tokens, never a hard-coded time', () => {
    expect(css).not.toMatch(/\b(animation|transition)[^;]*\b\d+(\.\d+)?m?s\b/);
  });
});

describe('touch target tokens', () => {
  it('keeps the minimum target at 44px (2.75rem at a 16px root)', () => {
    expect(lightTokens['--size-touch']).toBe('2.75rem');
  });

  it('never makes a key smaller than the minimum target', () => {
    expect(Number.parseFloat(lightTokens['--size-key'] ?? '0')).toBeGreaterThanOrEqual(2.75);
  });
});

describe('active operator', () => {
  const keyCss = entries.find(([path]) => path.endsWith('Key.module.css'))?.[1] ?? '';
  const active = rulesOf(keyCss).find((rule) => rule.selector.startsWith('.active,'));

  it('is drawn with a ring as well as a fill', () => {
    expect(active?.body).toMatch(/box-shadow:/);
    expect(active?.body).toMatch(/background:/);
  });

  it('keeps a ring in forced colours mode, where box-shadow is dropped', () => {
    expect(keyCss).toMatch(/@media \(forced-colors: active\)\s*\{\s*\.active\s*\{[^}]*outline:/);
  });
});
