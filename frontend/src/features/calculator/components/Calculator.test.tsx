import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { ApiClientContext, createApiClient, type CalculateRequest } from '../../../api';
import { expectNoAxeViolations } from '../../../test/axe';
import { server } from '../../../test/server';
import { Calculator } from './Calculator';

const BASE_URL = 'http://api.test';
const CALCULATE_URL = `${BASE_URL}/api/v1/calculate`;
const client = createApiClient({ baseUrl: BASE_URL });

/** Points the calculator at the mock server through the same context the app uses. */
function wrapper({ children }: { children: ReactNode }) {
  return <ApiClientContext value={client}>{children}</ApiClientContext>;
}

function renderCalculator() {
  render(<Calculator />, { wrapper });
  return userEvent.setup();
}

/** The whole display: expression and value. */
function display() {
  return screen.getByRole('group', { name: 'Display' });
}

/** The live region, which holds the value only. */
function liveRegion() {
  return screen.getByRole('status');
}

const OPERATIONS_URL = `${BASE_URL}/api/v1/operations`;

/** Answers the operations lookup with only the named operations. */
function reportOperations(...names: string[]) {
  server.use(
    http.get(OPERATIONS_URL, () =>
      HttpResponse.json({
        operations: names.map((name) => ({ name, arity: 2, symbol: name, description: name })),
      }),
    ),
  );
}

/** Clicks keys by their accessible names, in order. */
async function pressKeys(user: ReturnType<typeof userEvent.setup>, ...names: string[]) {
  for (const name of names) {
    await user.click(screen.getByRole('button', { name }));
  }
}

/** Records every request body and answers with the given response. */
function recordCalculate(respond: (body: CalculateRequest, attempt: number) => Response) {
  const bodies: CalculateRequest[] = [];
  server.use(
    http.post(CALCULATE_URL, async ({ request }) => {
      const body = (await request.json()) as CalculateRequest;
      bodies.push(body);
      return respond(body, bodies.length);
    }),
  );
  return bodies;
}

function divideByZero(): Response {
  return HttpResponse.json(
    {
      error: {
        code: 'DIVISION_BY_ZERO',
        message: 'Cannot divide by zero.',
        requestId: 'req-dz',
      },
    },
    { status: 422 },
  );
}

describe('Calculator', () => {
  it('adds 12.5 and 7.25 to get 19.75', async () => {
    const bodies = recordCalculate(() =>
      HttpResponse.json({ operation: 'add', operands: ['12.5', '7.25'], result: '19.75' }),
    );
    const user = renderCalculator();

    await pressKeys(user, '1', '2', 'Decimal point', '5', 'Add', '7', 'Decimal point', '2', '5');
    expect(display()).toHaveTextContent('12.5 +');
    await pressKeys(user, 'Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent('19.75');
    });
    expect(display()).toHaveTextContent('12.5 + 7.25 =');
    expect(bodies).toEqual([{ operation: 'add', operands: ['12.5', '7.25'] }]);
  });

  it('reads "20% of 50" from the % of key and sends x as the first operand', async () => {
    const bodies = recordCalculate(() =>
      HttpResponse.json({ operation: 'percentage', operands: ['20', '50'], result: '10' }),
    );
    const user = renderCalculator();

    await pressKeys(user, '2', '0', 'Percent of');

    expect(display()).toHaveTextContent('20% of');

    await pressKeys(user, '5', '0', 'Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent('20% of 50 =');
    });
    expect(liveRegion()).toHaveTextContent('10');
    expect(bodies).toEqual([{ operation: 'percentage', operands: ['20', '50'] }]);
  });

  it('shows friendly copy and no Retry when dividing by zero', async () => {
    recordCalculate(divideByZero);
    const user = renderCalculator();

    await pressKeys(user, '8', 'Divide', '0', 'Equals');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent("You can't divide by zero.");
    });
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('shows the rate limit message and no Retry when the server answers 429', async () => {
    recordCalculate(() =>
      HttpResponse.json(
        {
          error: {
            code: 'RATE_LIMITED',
            message: 'Too many requests. Please wait and try again.',
            requestId: 'req-rl',
            fields: { retryAfterSeconds: 2 },
          },
        },
        { status: 429, headers: { 'Retry-After': '2' } },
      ),
    );
    const user = renderCalculator();

    await pressKeys(user, '1', 'Add', '2', 'Equals');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Too many calculations too quickly. Wait a few seconds and try again.',
      );
    });
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('shows a loading state and disables equals while a request is pending', async () => {
    server.use(
      http.post(CALCULATE_URL, async () => {
        await delay(50);
        return HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' });
      }),
    );
    const user = renderCalculator();

    await pressKeys(user, '1', 'Add', '2', 'Equals');

    expect(liveRegion()).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByRole('button', { name: 'Equals' })).toBeDisabled();
    await waitFor(() => {
      expect(display()).toHaveTextContent('3');
    });
    expect(liveRegion()).toHaveAttribute('aria-busy', 'false');
    expect(screen.getByRole('button', { name: 'Equals' })).toBeEnabled();
  });

  it('retries the same request after a network failure and then succeeds', async () => {
    const bodies = recordCalculate((_body, attempt) =>
      attempt === 1
        ? HttpResponse.error()
        : HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' }),
    );
    const user = renderCalculator();

    await pressKeys(user, '1', 'Add', '2', 'Equals');
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent("Couldn't reach the server.");
    });
    await pressKeys(user, 'Retry');

    await waitFor(() => {
      expect(display()).toHaveTextContent('1 + 2 =');
    });
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
    });
    expect(display()).toHaveTextContent('3');
    expect(bodies).toEqual([
      { operation: 'add', operands: ['1', '2'] },
      { operation: 'add', operands: ['1', '2'] },
    ]);
  });

  it('shows the request ID for an error code it has no copy for', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.json(
          { error: { code: 'BRAND_NEW', message: 'raw', requestId: 'req-77' } },
          { status: 500 },
        ),
      ),
    );
    const user = renderCalculator();

    await pressKeys(user, '1', 'Add', '2', 'Equals');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Something went wrong. Request ID: req-77',
      );
    });
  });

  it('runs a full calculation from the keyboard alone', async () => {
    const bodies = recordCalculate(() =>
      HttpResponse.json({ operation: 'multiply', operands: ['12.5', '4'], result: '50' }),
    );
    const user = renderCalculator();

    await user.keyboard('12.5*4{Enter}');

    await waitFor(() => {
      expect(display()).toHaveTextContent('50');
    });
    expect(bodies).toEqual([{ operation: 'multiply', operands: ['12.5', '4'] }]);
  });

  it('edits and clears from the keyboard', async () => {
    const user = renderCalculator();

    await user.keyboard('123{Backspace}');
    expect(display()).toHaveTextContent('12');
    await user.keyboard('{Escape}');

    expect(display()).toHaveTextContent('0');
    expect(display()).not.toHaveTextContent('12');
  });

  it('takes a square root with the r key', async () => {
    const bodies = recordCalculate(() =>
      HttpResponse.json({ operation: 'sqrt', operands: ['144'], result: '12' }),
    );
    const user = renderCalculator();

    await user.keyboard('144r');

    await waitFor(() => {
      expect(display()).toHaveTextContent('12');
    });
    expect(bodies).toEqual([{ operation: 'sqrt', operands: ['144'] }]);
  });

  it('reuses a result from the history in the next calculation', async () => {
    const bodies = recordCalculate((body) =>
      body.operation === 'add'
        ? HttpResponse.json({ operation: 'add', operands: ['12.5', '7.25'], result: '19.75' })
        : HttpResponse.json({ operation: 'multiply', operands: body.operands, result: '39.5' }),
    );
    const user = renderCalculator();
    await user.keyboard('12.5+7.25{Enter}');
    await waitFor(() => {
      expect(within(screen.getByRole('list')).getAllByRole('listitem')).toHaveLength(1);
    });
    await pressKeys(user, 'Clear');
    expect(display()).not.toHaveTextContent('19.75');

    await user.click(screen.getByRole('button', { name: /12\.5 \+ 7\.25/ }));
    expect(display()).toHaveTextContent('19.75');
    await pressKeys(user, 'Multiply', '2', 'Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent('39.5');
    });
    expect(bodies.at(-1)).toEqual({ operation: 'multiply', operands: ['19.75', '2'] });
  });

  it('runs equals with Enter after reusing a history entry with the mouse', async () => {
    const bodies = recordCalculate((body) =>
      body.operation === 'add' && body.operands[0] === '2'
        ? HttpResponse.json({ operation: 'add', operands: ['2', '3'], result: '5' })
        : HttpResponse.json({ operation: 'add', operands: body.operands, result: '6' }),
    );
    const user = renderCalculator();
    await user.keyboard('2+3{Enter}');
    await waitFor(() => {
      expect(within(screen.getByRole('list')).getAllByRole('listitem')).toHaveLength(1);
    });
    await user.keyboard('{Escape}');

    await user.click(screen.getByRole('button', { name: /2 \+ 3/ }));
    await user.keyboard('+1{Enter}');

    await waitFor(() => {
      expect(display()).toHaveTextContent('6');
    });
    expect(bodies.at(-1)).toEqual({ operation: 'add', operands: ['5', '1'] });
  });

  it('shows which operator is waiting for its second operand', async () => {
    const user = renderCalculator();

    await pressKeys(user, '5', 'Add');
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'true');

    await pressKeys(user, 'Clear');
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('keeps Escape inside the open history drawer from clearing the display', async () => {
    const user = renderCalculator();
    await pressKeys(user, '7');
    const toggle = screen.getByRole('button', { name: 'History' });
    await user.click(toggle);

    await user.keyboard('{Escape}');

    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(display()).toHaveTextContent('7');
  });

  it('announces an error as an alert and leaves the status region for the display', async () => {
    recordCalculate(divideByZero);
    const user = renderCalculator();

    await pressKeys(user, '8', 'Divide', '0', 'Equals');

    expect(await screen.findByRole('alert')).toHaveTextContent("You can't divide by zero.");
    expect(display()).not.toContainElement(screen.getByRole('alert'));
  });

  it('moves focus to the display after Retry, because the Retry button disappears', async () => {
    recordCalculate((_body, attempt) =>
      attempt === 1
        ? HttpResponse.error()
        : HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' }),
    );
    const user = renderCalculator();
    await pressKeys(user, '1', 'Add', '2', 'Equals');
    const retry = await screen.findByRole('button', { name: 'Retry' });
    retry.focus();

    await user.keyboard('{Enter}');

    expect(display()).toHaveFocus();
    await waitFor(() => {
      expect(liveRegion()).toHaveTextContent('3');
    });
    expect(display()).toHaveFocus();
  });

  it('puts only the value in the live region', async () => {
    const user = renderCalculator();

    await pressKeys(user, '1', 'Add');

    expect(display()).toHaveTextContent('1 +');
    expect(liveRegion()).not.toHaveTextContent('+');
  });

  describe('operations reported by the backend', () => {
    it('disables the key and the shortcut for an operation the server does not report', async () => {
      reportOperations('add', 'subtract', 'multiply', 'divide', 'percentage', 'sqrt');
      const user = renderCalculator();
      const power = screen.getByRole('button', { name: 'Power' });
      await waitFor(() => {
        expect(power).toHaveAttribute('aria-disabled', 'true');
      });
      expect(power).toHaveAttribute('title', 'Not available on this server');
      expect(screen.getByRole('button', { name: 'Add' })).not.toHaveAttribute('aria-disabled');

      await user.click(power);
      await user.keyboard('^');

      expect(power).toHaveAttribute('aria-pressed', 'false');
      expect(display()).not.toHaveTextContent('^');
      await user.keyboard('5+');
      expect(display()).toHaveTextContent('5 +');
    });

    it('ignores the square root shortcut when sqrt is not reported', async () => {
      reportOperations('add');
      const user = renderCalculator();
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Square root' })).toHaveAttribute(
          'aria-disabled',
          'true',
        );
      });
      const bodies = recordCalculate(() =>
        HttpResponse.json({ operation: 'sqrt', operands: ['9'], result: '3' }),
      );

      await user.keyboard('9r');

      expect(display()).toHaveTextContent('9');
      expect(bodies).toEqual([]);
    });

    it('renders every key at once, before the list arrives', () => {
      server.use(http.get(OPERATIONS_URL, () => delay('infinite')));
      renderCalculator();

      expect(screen.getByRole('button', { name: 'Power' })).not.toHaveAttribute('aria-disabled');
      expect(screen.getByRole('button', { name: 'Equals' })).toBeEnabled();
    });

    it('leaves every key enabled and shows no error when the lookup fails', async () => {
      server.use(http.get(OPERATIONS_URL, () => HttpResponse.error()));
      const user = renderCalculator();

      await user.keyboard('2^');

      expect(display()).toHaveTextContent('2 ^');
      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
      expect(
        screen.getAllByRole('button').filter((key) => key.getAttribute('aria-disabled') === 'true'),
      ).toEqual([]);
    });

    it('leaves every key enabled when the server reports no operations at all', async () => {
      let asked = false;
      server.use(
        http.get(OPERATIONS_URL, () => {
          asked = true;
          return HttpResponse.json({ operations: [] });
        }),
      );
      const user = renderCalculator();
      await waitFor(() => {
        expect(asked).toBe(true);
      });

      await user.keyboard('2^');

      expect(display()).toHaveTextContent('2 ^');
    });
  });

  it('keeps a result that is too long to reuse readable and says why it is refused', async () => {
    const long = '9'.repeat(65);
    const bodies = recordCalculate(() =>
      HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: long }),
    );
    const user = renderCalculator();
    await pressKeys(user, '1', 'Add', '2', 'Equals');
    await waitFor(() => {
      expect(liveRegion()).toHaveTextContent(long);
    });

    await pressKeys(user, 'Add');

    expect(screen.getByRole('alert')).toHaveTextContent(
      'This result is too long to use in another calculation.',
    );
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
    expect(liveRegion()).toHaveTextContent(long);
    expect(bodies).toHaveLength(1);
    await pressKeys(user, '4');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  describe('axe', () => {
    it('finds no violations when idle', async () => {
      const { container } = render(<Calculator />, { wrapper });

      await expectNoAxeViolations(container);
    });

    it('finds no violations when showing a result', async () => {
      recordCalculate(() =>
        HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' }),
      );
      const { container } = render(<Calculator />, { wrapper });
      const user = userEvent.setup();

      await pressKeys(user, '1', 'Add', '2', 'Equals');
      await waitFor(() => {
        expect(display()).toHaveTextContent('3');
      });

      await expectNoAxeViolations(container);
    });

    it('finds no violations while an operator is active', async () => {
      const { container } = render(<Calculator />, { wrapper });
      const user = userEvent.setup();

      await pressKeys(user, '1', 'Add');

      await expectNoAxeViolations(container);
    });

    it('finds no violations in the error state', async () => {
      recordCalculate(divideByZero);
      const { container } = render(<Calculator />, { wrapper });
      const user = userEvent.setup();

      await pressKeys(user, '8', 'Divide', '0', 'Equals');
      await screen.findByRole('alert');

      await expectNoAxeViolations(container);
    });

    it('finds no violations with a retryable error and its Retry button', async () => {
      recordCalculate(() => HttpResponse.error());
      const { container } = render(<Calculator />, { wrapper });
      const user = userEvent.setup();

      await pressKeys(user, '1', 'Add', '2', 'Equals');
      await screen.findByRole('button', { name: 'Retry' });

      await expectNoAxeViolations(container);
    });

    it('finds no violations with the history drawer open and holding entries', async () => {
      recordCalculate(() =>
        HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' }),
      );
      const { container } = render(<Calculator />, { wrapper });
      const user = userEvent.setup();
      await pressKeys(user, '1', 'Add', '2', 'Equals');
      await screen.findByRole('list');

      await user.click(screen.getByRole('button', { name: 'History' }));

      expect(screen.getByRole('button', { name: 'History' })).toHaveAttribute(
        'aria-expanded',
        'true',
      );
      await expectNoAxeViolations(container);
    });
  });

  it('reads Error, not a number, after dividing by zero, and a digit starts fresh', async () => {
    recordCalculate(divideByZero);
    const user = renderCalculator();

    await pressKeys(user, '9', 'Divide', '0', 'Equals');
    await screen.findByRole('alert');

    expect(display()).toHaveTextContent('Error');
    expect(display()).not.toHaveTextContent('=');
    await pressKeys(user, '4');
    expect(display()).toHaveTextContent('4');
    expect(display()).not.toHaveTextContent('Error');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});

describe('Calculator server limits', () => {
  function reportLimits(limits: unknown) {
    server.use(
      http.get(OPERATIONS_URL, () =>
        HttpResponse.json({
          operations: [{ name: 'add', arity: 2, symbol: '+', description: 'add' }],
          ...(limits === undefined ? {} : { limits }),
        }),
      ),
    );
  }

  async function typeDigits(count: number) {
    const user = renderCalculator();
    // The limit arrives with the operations lookup, so wait for the keypad to apply it.
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Power' })).toHaveAttribute(
        'aria-disabled',
        'true',
      );
    });
    for (let index = 0; index < count; index += 1) {
      await user.click(screen.getByRole('button', { name: '7' }));
    }
  }

  it('stops typing at the operand length the server reports', async () => {
    reportLimits({ maxOperandLength: 5, maxExponent: 100, decimalPrecision: 10 });

    await typeDigits(8);

    expect(liveRegion()).toHaveTextContent(/^77777$/);
  });

  it('falls back to the built-in 64 when the server sends no limits', async () => {
    reportLimits(undefined);

    await typeDigits(66);

    expect(liveRegion().textContent.replace(/\D/g, '')).toHaveLength(64);
  });

  it('falls back to the built-in 64 when the operations request fails', async () => {
    server.use(http.get(OPERATIONS_URL, () => HttpResponse.error()));
    const user = renderCalculator();

    for (let index = 0; index < 66; index += 1) {
      await user.click(screen.getByRole('button', { name: '7' }));
    }

    expect(liveRegion().textContent.replace(/\D/g, '')).toHaveLength(64);
  });
});
