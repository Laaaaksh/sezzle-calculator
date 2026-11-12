import { act, renderHook, waitFor } from '@testing-library/react';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { createApiClient } from '../../api';
import type { CalculateRequest, CalculateResponse } from '../../api';
import { DIGITS, HISTORY_LIMIT, type BinaryOperation } from '../../constants/calculator';
import { server } from '../../test/server';
import type { CalculatorAction, CalculatorState } from './state';
import { useCalculator, type CalculatorClient } from './useCalculator';

const BASE_URL = 'http://api.test';
const CALCULATE_URL = `${BASE_URL}/api/v1/calculate`;
const client = createApiClient({ baseUrl: BASE_URL });

function digitsOf(text: string): CalculatorAction[] {
  const actions: CalculatorAction[] = [];
  for (const char of text) {
    const digit = DIGITS.find((candidate) => candidate === char);
    actions.push(char === '.' ? { type: 'decimal' } : { type: 'digit', digit: digit ?? '0' });
  }
  return actions;
}

function op(operator: BinaryOperation): CalculatorAction {
  return { type: 'operator', operator };
}

/** Presses "a op b" without pressing equals. */
function expression(a: string, operator: BinaryOperation, b: string): CalculatorAction[] {
  return [...digitsOf(a), op(operator), ...digitsOf(b)];
}

const EQUALS: CalculatorAction = { type: 'equals' };

function press(
  result: { current: { dispatch: (action: CalculatorAction) => void } },
  ...actions: CalculatorAction[]
): void {
  act(() => {
    for (const action of actions) {
      result.current.dispatch(action);
    }
  });
}

function addHandler(): void {
  server.use(
    http.post(CALCULATE_URL, async ({ request }) => {
      const body = (await request.json()) as CalculateRequest;
      return HttpResponse.json({
        operation: body.operation,
        operands: body.operands,
        result: 'ok',
      } satisfies CalculateResponse);
    }),
  );
}

/** A client whose answers the test releases by hand, and which ignores the abort signal. */
function manualClient() {
  const calls: { request: CalculateRequest; respond: (result: string) => void }[] = [];
  const fake: CalculatorClient = {
    calculate: (request) =>
      new Promise<CalculateResponse>((resolve) => {
        calls.push({
          request,
          respond: (result) => {
            resolve({ operation: request.operation, operands: request.operands, result });
          },
        });
      }),
  };
  return { fake, calls };
}

function statusOf(result: { current: { state: CalculatorState } }): string {
  return result.current.state.status;
}

describe('useCalculator', () => {
  it('sends the operands as strings and shows the result', async () => {
    let seenBody: unknown;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        seenBody = await request.json();
        return HttpResponse.json({
          operation: 'add',
          operands: ['12.5', '7.25'],
          result: '19.75',
        });
      }),
    );
    const { result } = renderHook(() => useCalculator({ client }));

    press(result, ...expression('12.5', 'add', '7.25'), EQUALS);

    await waitFor(() => {
      expect(statusOf(result)).toBe('showing_result');
    });
    expect(seenBody).toEqual({ operation: 'add', operands: ['12.5', '7.25'] });
    expect(result.current.state).toMatchObject({ result: '19.75' });
  });

  it('reports pending while the request is in flight', async () => {
    const { fake, calls } = manualClient();
    const { result } = renderHook(() => useCalculator({ client: fake }));

    press(result, ...expression('1', 'add', '2'), EQUALS);

    expect(result.current.isPending).toBe(true);
    act(() => calls[0]?.respond('3'));
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
  });

  it('sends one request when equals is pressed twice while pending', async () => {
    let requests = 0;
    server.use(
      http.post(CALCULATE_URL, async () => {
        requests += 1;
        await delay(20);
        return HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' });
      }),
    );
    const { result } = renderHook(() => useCalculator({ client }));

    press(result, ...expression('1', 'add', '2'), EQUALS, EQUALS, EQUALS);

    await waitFor(() => {
      expect(statusOf(result)).toBe('showing_result');
    });
    expect(requests).toBe(1);
  });

  it('cancels the in-flight request when the user clears', async () => {
    let aborted = false;
    let started = false;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        started = true;
        request.signal.addEventListener('abort', () => {
          aborted = true;
        });
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );
    const { result } = renderHook(() => useCalculator({ client }));
    press(result, ...expression('1', 'add', '2'), EQUALS);
    await waitFor(() => {
      expect(started).toBe(true);
    });

    press(result, { type: 'clear' });

    await waitFor(() => {
      expect(aborted).toBe(true);
    });
    expect(result.current.state).toMatchObject({ status: 'entering_first', operand: '' });
    expect(result.current.isPending).toBe(false);
  });

  it('does not show a canceled request as an error', async () => {
    server.use(http.post(CALCULATE_URL, () => delay('infinite')));
    const { result } = renderHook(() => useCalculator({ client }));
    press(result, ...expression('1', 'add', '2'), EQUALS);

    press(result, { type: 'clear' });
    await act(async () => {
      await delay(20);
    });

    expect(statusOf(result)).toBe('entering_first');
  });

  it('aborts the request when the component unmounts', async () => {
    let aborted = false;
    let started = false;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        started = true;
        request.signal.addEventListener('abort', () => {
          aborted = true;
        });
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );
    const { result, unmount } = renderHook(() => useCalculator({ client }));
    press(result, ...expression('1', 'add', '2'), EQUALS);
    await waitFor(() => {
      expect(started).toBe(true);
    });

    unmount();

    await waitFor(() => {
      expect(aborted).toBe(true);
    });
  });

  it('ignores a stale response to a request the user cleared', async () => {
    const { fake, calls } = manualClient();
    const { result } = renderHook(() => useCalculator({ client: fake }));
    press(result, ...expression('1', 'add', '2'), EQUALS);
    press(result, { type: 'clear' }, ...expression('5', 'add', '5'), EQUALS);

    act(() => calls[1]?.respond('10'));
    await waitFor(() => {
      expect(result.current.state).toMatchObject({ status: 'showing_result', result: '10' });
    });
    act(() => calls[0]?.respond('3'));

    expect(result.current.state).toMatchObject({ result: '10' });
    expect(result.current.history.map((entry) => entry.result)).toEqual(['10']);
  });

  it('shows the error code and request ID when the server refuses', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.json(
          {
            error: {
              code: 'DIVISION_BY_ZERO',
              message: 'Cannot divide by zero.',
              requestId: 'req-42',
            },
          },
          { status: 422 },
        ),
      ),
    );
    const { result } = renderHook(() => useCalculator({ client }));

    press(result, ...expression('1', 'divide', '0'), EQUALS);

    await waitFor(() => {
      expect(statusOf(result)).toBe('error');
    });
    expect(result.current.state).toMatchObject({
      error: { code: 'DIVISION_BY_ZERO', requestId: 'req-42', retryable: false },
    });
  });

  it('does not add a failed calculation to the history', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.error()));
    const { result } = renderHook(() => useCalculator({ client }));

    press(result, ...expression('1', 'add', '2'), EQUALS);
    await waitFor(() => {
      expect(statusOf(result)).toBe('error');
    });

    expect(result.current.history).toEqual([]);
  });

  it('retries the same request after a network failure', async () => {
    const bodies: unknown[] = [];
    let attempt = 0;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        bodies.push(await request.json());
        attempt += 1;
        return attempt === 1
          ? HttpResponse.error()
          : HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: '3' });
      }),
    );
    const { result } = renderHook(() => useCalculator({ client }));
    press(result, ...expression('1', 'add', '2'), EQUALS);
    await waitFor(() => {
      expect(result.current.state).toMatchObject({ status: 'error', error: { retryable: true } });
    });

    press(result, { type: 'retry' });

    await waitFor(() => {
      expect(result.current.state).toMatchObject({ status: 'showing_result', result: '3' });
    });
    expect(bodies).toEqual([
      { operation: 'add', operands: ['1', '2'] },
      { operation: 'add', operands: ['1', '2'] },
    ]);
  });
});

describe('history', () => {
  it('records the expression and the result of each calculation', async () => {
    addHandler();
    const { result } = renderHook(() => useCalculator({ client }));

    press(result, ...expression('12', 'add', '7'), EQUALS);
    await waitFor(() => {
      expect(result.current.history).toHaveLength(1);
    });

    expect(result.current.history[0]).toMatchObject({
      request: { operation: 'add', operands: ['12', '7'] },
      result: 'ok',
    });
  });

  it('puts the newest entry on top', async () => {
    const { fake, calls } = manualClient();
    const { result } = renderHook(() => useCalculator({ client: fake }));
    press(result, ...expression('1', 'add', '1'), EQUALS);
    act(() => calls[0]?.respond('2'));
    await waitFor(() => {
      expect(statusOf(result)).toBe('showing_result');
    });
    press(result, ...digitsOf('3'), op('add'), ...digitsOf('3'), EQUALS);
    act(() => calls[1]?.respond('6'));
    await waitFor(() => {
      expect(result.current.history).toHaveLength(2);
    });

    expect(result.current.history.map((entry) => entry.result)).toEqual(['6', '2']);
  });

  it('keeps at most 20 entries and drops the oldest', async () => {
    const calculate = vi.fn<CalculatorClient['calculate']>((request) =>
      Promise.resolve({
        operation: request.operation,
        operands: request.operands,
        result: `r${request.operands[0] ?? ''}`,
      }),
    );
    const { result } = renderHook(() => useCalculator({ client: { calculate } }));

    for (let n = 1; n <= HISTORY_LIMIT + 1; n += 1) {
      press(result, { type: 'clear' }, ...digitsOf(String(n)), op('add'), ...digitsOf('1'), EQUALS);
      await waitFor(() => {
        expect(statusOf(result)).toBe('showing_result');
      });
    }

    expect(result.current.history).toHaveLength(HISTORY_LIMIT);
    expect(result.current.history[0]?.result).toBe(`r${HISTORY_LIMIT + 1}`);
    expect(result.current.history.at(-1)?.result).toBe('r2');
  });

  it('reuses the result of an entry when it is clicked', async () => {
    addHandler();
    const { result } = renderHook(() => useCalculator({ client }));
    press(result, ...expression('12', 'add', '7'), EQUALS);
    await waitFor(() => {
      expect(result.current.history).toHaveLength(1);
    });
    press(result, { type: 'clear' });
    const entry = result.current.history[0];
    if (entry === undefined) {
      throw new Error('Expected a history entry');
    }

    act(() => {
      result.current.reuse(entry);
    });

    expect(result.current.state).toMatchObject({
      status: 'showing_result',
      result: 'ok',
      request: null,
    });
  });
});
