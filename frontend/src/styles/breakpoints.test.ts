import postcss, { type AtRule } from 'postcss';
import { describe, expect, it } from 'vitest';
import { WIDE_LAYOUT_MIN_WIDTH_PX, WIDE_LAYOUT_QUERY } from '../constants/layout';

// The ?raw glob reads every stylesheet, so a new module that repeats the breakpoint is checked too.
const stylesheets = import.meta.glob<string>('../**/*.css', {
  query: '?raw',
  import: 'default',
  eager: true,
});

/** The sub-pixel step CSS uses for "just below" a min-width breakpoint. */
const BELOW_STEP_PX = 0.02;

type WidthFeature = 'min-width' | 'max-width';

interface WidthCondition {
  file: string;
  feature: WidthFeature;
  value: string;
}

const widthConditions: WidthCondition[] = Object.entries(stylesheets).flatMap(([file, css]) => {
  const conditions: WidthCondition[] = [];
  postcss.parse(css, { from: file }).walkAtRules('media', (rule: AtRule) => {
    for (const match of rule.params.matchAll(/\(\s*(min-width|max-width)\s*:\s*([^)]+?)\s*\)/gi)) {
      const [, feature, value] = match;
      if (feature !== undefined && value !== undefined) {
        conditions.push({ file, feature: feature.toLowerCase() as WidthFeature, value });
      }
    }
  });
  return conditions;
});

/** A width in px. Any other unit fails, so a breakpoint cannot hide behind em or rem. */
function pixels({ file, value }: WidthCondition): number {
  const match = /^(\d*\.?\d+)px$/i.exec(value);
  if (match?.[1] === undefined) {
    throw new Error(`${file}: breakpoint "${value}" is not in px`);
  }
  return Number(match[1]);
}

function conditionsFor(feature: WidthFeature): WidthCondition[] {
  return widthConditions.filter((condition) => condition.feature === feature);
}

describe('wide layout breakpoint', () => {
  it('builds the matchMedia query from the one constant', () => {
    expect(WIDE_LAYOUT_QUERY).toBe(`(min-width: ${WIDE_LAYOUT_MIN_WIDTH_PX}px)`);
  });

  it('is the only min-width breakpoint in the stylesheets, repeated verbatim', () => {
    const found = conditionsFor('min-width');

    expect(found.length).toBeGreaterThanOrEqual(3);
    for (const condition of found) {
      expect(pixels(condition), condition.file).toBe(WIDE_LAYOUT_MIN_WIDTH_PX);
    }
  });

  it('pairs every max-width rule with the same breakpoint, one step below', () => {
    const below = conditionsFor('max-width');

    expect(below.length).toBeGreaterThan(0);
    for (const condition of below) {
      expect(pixels(condition), condition.file).toBeCloseTo(
        WIDE_LAYOUT_MIN_WIDTH_PX - BELOW_STEP_PX,
        5,
      );
    }
  });
});
