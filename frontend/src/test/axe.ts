import axe, { type Result } from 'axe-core';

function describeViolation(violation: Result): string {
  const targets = violation.nodes.map((node) => node.target.join(' ')).join(', ');
  return `${violation.id} (${violation.impact ?? 'unknown'}): ${violation.help} -> ${targets}`;
}

/**
 * Fails the test when axe finds an accessibility violation inside `container`.
 *
 * The color-contrast rule is switched off here and only here. jsdom does not do layout or
 * resolve stylesheets, so axe cannot read real colours and the rule gives false results.
 * Contrast is still enforced: contrast.test.ts computes the WCAG ratios from the values in
 * tokens.css for both themes. Do not turn this exception into a wider one.
 *
 * It throws a plain Error so the rule ids show up in the failure message.
 */
export async function expectNoAxeViolations(container: Element): Promise<void> {
  const results = await axe.run(container, {
    resultTypes: ['violations'],
    rules: { 'color-contrast': { enabled: false } },
  });

  if (results.violations.length > 0) {
    throw new Error(
      `axe found violations:\n${results.violations.map(describeViolation).join('\n')}`,
    );
  }
}
