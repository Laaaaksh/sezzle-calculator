import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { createApiClient, type CalculateRequest } from '../../../api';
import { server } from '../../../test/server';
import { Calculator } from './Calculator';

const BASE_URL = 'http://api.test';
const CALCULATE_URL = `${BASE_URL}/api/v1/calculate`;
const client = createApiClient({ baseUrl: BASE_URL });

function renderCalculator() {
  render(<Calculator client={client} />);
  return userEvent.setup();
}

function display() {
  return screen.getByRole('status');
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

  it('shows friendly copy and no Retry when dividing by zero', async () => {
    recordCalculate(divideByZero);
    const user = renderCalculator();

    await pressKeys(user, '8', 'Divide', '0', 'Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent("You can't divide by zero.");
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

    expect(display()).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByRole('button', { name: 'Equals' })).toBeDisabled();
    await waitFor(() => {
      expect(display()).toHaveTextContent('3');
    });
    expect(display()).toHaveAttribute('aria-busy', 'false');
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
      expect(display()).toHaveTextContent("Couldn't reach the server.");
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
      expect(display()).toHaveTextContent('Something went wrong. Request ID: req-77');
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
});
