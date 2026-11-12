import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from './axe';

describe('expectNoAxeViolations', () => {
  it('passes for markup with no problems', async () => {
    const { container } = render(<button type="button">Save</button>);

    await expectNoAxeViolations(container);
  });

  it('fails and names the rule when a button has no accessible name', async () => {
    const { container } = render(<button type="button" />);

    await expect(expectNoAxeViolations(container)).rejects.toThrow(/button-name/);
  });
});
