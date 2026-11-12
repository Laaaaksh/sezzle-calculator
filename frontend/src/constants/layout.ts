/**
 * Width in px at which the history tape moves beside the keypad. CSS cannot read a TypeScript
 * constant, so every `min-width: 900px` media query in the stylesheets repeats this number;
 * styles/breakpoints.test.ts fails if any of them drifts from it. Change all of them together.
 */
export const WIDE_LAYOUT_MIN_WIDTH_PX = 900;

/** The matchMedia query for the wide layout, built from the one constant above. */
export const WIDE_LAYOUT_QUERY = `(min-width: ${WIDE_LAYOUT_MIN_WIDTH_PX}px)`;
