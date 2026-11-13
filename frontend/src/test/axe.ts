import axe, { type Result } from 'axe-core';

function describeViolation(violation: Result): string {
  const targets = violation.nodes.map((node) => node.target.join(' ')).join(', ');
  return `${violation.id} (${violation.impact ?? 'unknown'}): ${violation.help} -> ${targets}`;
}

/**
 * Fails when axe finds an accessibility violation inside `container`.
 * The color-contrast rule is off: jsdom does no layout, so axe reads no real colours. Contrast is
 * still enforced by contrast.test.ts from tokens.css. Do not widen this exception.
 * Throws a plain Error so the rule ids appear in the failure message.
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
