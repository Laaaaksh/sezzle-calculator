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

describe('wide layout breakpoint', () => {
  it('builds the matchMedia query from the one constant', () => {
    expect(WIDE_LAYOUT_QUERY).toBe(`(min-width: ${WIDE_LAYOUT_MIN_WIDTH_PX}px)`);
  });

  it('is the only min-width breakpoint in the stylesheets, repeated verbatim', () => {
    const found = Object.entries(stylesheets).flatMap(([file, css]) =>
      [...css.matchAll(/@media[^{]*\(min-width:\s*([\d.]+)px\)/g)].map((match) => ({
        file,
        width: Number(match[1]),
      })),
    );

    expect(found.length).toBeGreaterThanOrEqual(3);
    for (const { file, width } of found) {
      expect(width, file).toBe(WIDE_LAYOUT_MIN_WIDTH_PX);
    }
  });

  it('pairs every max-width rule with the same breakpoint, one step below', () => {
    const below = Object.entries(stylesheets).flatMap(([file, css]) =>
      [...css.matchAll(/\(max-width:\s*([\d.]+)px\)/g)].map((match) => ({
        file,
        width: Number(match[1]),
      })),
    );

    expect(below.length).toBeGreaterThan(0);
    for (const { file, width } of below) {
      expect(width, file).toBeCloseTo(WIDE_LAYOUT_MIN_WIDTH_PX - BELOW_STEP_PX, 5);
    }
  });
});
