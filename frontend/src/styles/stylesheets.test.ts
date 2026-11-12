import postcss, {
  type AtRule,
  type Container,
  type Declaration,
  type Node,
  type Root,
  type Rule,
} from 'postcss';
import { describe, expect, it } from 'vitest';
import { lightTokens } from '../test/tokens';

/** Every CSS module in the app, parsed. Written so a new component is checked without edits. */
const modules = import.meta.glob<string>('../**/*.module.css', {
  query: '?raw',
  import: 'default',
  eager: true,
});
const entries: [string, Root][] = Object.entries(modules).map(([path, css]) => [
  path,
  postcss.parse(css, { from: path }),
]);

const TIME = /^-?\d*\.?\d+m?s$/i;

function rulesOf(root: Root): Rule[] {
  const rules: Rule[] = [];
  root.walkRules((rule) => {
    if (rule.parent?.type !== 'atrule' || (rule.parent as AtRule).name !== 'keyframes') {
      rules.push(rule);
    }
  });
  return rules;
}

function declsOf(container: Container, prop: string): Declaration[] {
  return (container.nodes ?? []).filter(
    (node): node is Declaration => node.type === 'decl' && node.prop === prop,
  );
}

function hasDecl(rule: Rule, prop: string, value?: string): boolean {
  return declsOf(rule, prop).some((decl) => value === undefined || decl.value.trim() === value);
}

function isClickable(rule: Rule): boolean {
  return hasDecl(rule, 'cursor', 'pointer');
}

/** Media queries wrapping a node, normalised to lowercase with no spaces. */
function mediaQueries(node: Rule | Declaration): string[] {
  const queries: string[] = [];
  let parent: Node | undefined = node.parent;
  while (parent !== undefined) {
    if (parent.type === 'atrule' && (parent as AtRule).name === 'media') {
      queries.push((parent as AtRule).params.replace(/\s+/g, '').toLowerCase());
    }
    parent = parent.parent;
  }
  return queries;
}

function inMedia(node: Rule | Declaration, feature: string): boolean {
  const wanted = feature.replace(/\s+/g, '').toLowerCase();
  return mediaQueries(node).some((query) => query.includes(wanted));
}

/** Outline width in px, or undefined when the declaration has no px width. */
function outlineWidth(value: string): number | undefined {
  const width = postcss.list.space(value).find((part) => /^\d*\.?\d+px$/.test(part));
  return width === undefined ? undefined : Number.parseFloat(width);
}

function motionDecls(root: Root): Declaration[] {
  const decls: Declaration[] = [];
  root.walkDecls(/^(animation|transition)(-.+)?$/, (decl) => {
    decls.push(decl);
  });
  return decls;
}

describe('the CSS modules were found', () => {
  it('includes the calculator keys', () => {
    expect(entries.some(([path]) => path.endsWith('Key.module.css'))).toBe(true);
  });
});

describe.each(entries)('%s', (_path, root) => {
  it('never removes the focus outline', () => {
    root.walkDecls(/^outline(-style|-width)?$/, (decl) => {
      const parts = postcss.list.space(decl.value.toLowerCase());
      expect(parts, decl.parent?.toString()).not.toContain('none');
      expect(parts, decl.parent?.toString()).not.toContain('0');
    });
  });

  it('gives every clickable thing a visible focus ring of at least 2px', () => {
    const rules = rulesOf(root);
    if (!rules.some(isClickable)) {
      return;
    }
    const focusRules = rules.filter((rule) =>
      rule.selectors.some((selector) => selector.includes(':focus-visible')),
    );
    expect(focusRules.length).toBeGreaterThan(0);
    for (const rule of focusRules) {
      for (const decl of declsOf(rule, 'outline')) {
        const width = outlineWidth(decl.value);
        // A rule that only moves the offset (the active operator) leans on the base ring.
        if (width !== undefined) {
          expect(width, rule.selector).toBeGreaterThanOrEqual(2);
        }
      }
    }
  });

  it('makes every clickable rule at least 44px tall', () => {
    for (const rule of rulesOf(root).filter(isClickable)) {
      const heights = declsOf(rule, 'min-height').map((decl) => decl.value.trim());
      expect(heights, rule.selector).toSatisfy((values: string[]) =>
        values.some((value) => value === 'var(--size-touch)' || value === 'var(--size-key)'),
      );
    }
  });

  it('keeps animation and transitions inside the no-preference reduced motion query', () => {
    for (const decl of motionDecls(root)) {
      expect(inMedia(decl, 'prefers-reduced-motion: no-preference'), decl.toString()).toBe(true);
    }
  });

  it('takes durations from tokens, never a hard-coded time', () => {
    for (const decl of motionDecls(root)) {
      const parts = postcss.list.comma(decl.value).flatMap((part) => postcss.list.space(part));
      expect(
        parts.filter((part) => TIME.test(part)),
        decl.toString(),
      ).toEqual([]);
    }
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

function moduleNamed(name: string): Root {
  const root = entries.find(([path]) => path.endsWith(name))?.[1];
  if (root === undefined) {
    throw new Error(`no CSS module named ${name}`);
  }
  return root;
}

function rulesFor(root: Root, selector: string): Rule[] {
  return rulesOf(root).filter((rule) => rule.selectors.includes(selector));
}

describe('active operator', () => {
  const keyRules = rulesFor(moduleNamed('Key.module.css'), '.active');

  it('is drawn with a ring as well as a fill', () => {
    const base = keyRules.filter((rule) => mediaQueries(rule).length === 0);

    expect(base.some((rule) => hasDecl(rule, 'box-shadow'))).toBe(true);
    expect(base.some((rule) => hasDecl(rule, 'background'))).toBe(true);
  });

  it('keeps a ring in forced colours mode, where box-shadow is dropped', () => {
    const forced = keyRules.filter((rule) => inMedia(rule, 'forced-colors: active'));

    expect(
      forced.some((rule) =>
        declsOf(rule, 'outline').some((decl) => (outlineWidth(decl.value) ?? 0) >= 2),
      ),
    ).toBe(true);
  });
});

describe('history heading', () => {
  const drawer = moduleNamed('HistoryDrawer.module.css');

  it('is visually hidden in the drawer, where the toggle already says History', () => {
    const hidden = rulesFor(drawer, '.panel h2').filter((rule) =>
      mediaQueries(rule).some((query) => query.includes('max-width')),
    );

    expect(hidden.some((rule) => hasDecl(rule, 'clip-path', 'inset(50%)'))).toBe(true);
    expect(hidden.some((rule) => hasDecl(rule, 'position', 'absolute'))).toBe(true);
  });

  it('shows the toggle only where the heading is hidden, so History is always labelled once', () => {
    const wideToggle = rulesFor(drawer, '.toggle').filter((rule) =>
      mediaQueries(rule).some((query) => query.includes('min-width:900px')),
    );

    expect(wideToggle.some((rule) => hasDecl(rule, 'display', 'none'))).toBe(true);
  });
});
