import { delay, http, HttpResponse } from 'msw';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { server } from '../test/server';
import { createApiClient, resolveBaseUrl } from './client';
import { ApiError, CanceledError, NetworkError, TimeoutError } from './errors';
import { ERROR_CODES, type ApiErrorBody, type OperationInfo } from './types';

const BASE_URL = 'http://api.test';
const CALCULATE_URL = `${BASE_URL}/api/v1/calculate`;
const OPERATIONS_URL = `${BASE_URL}/api/v1/operations`;

const client = createApiClient({ baseUrl: BASE_URL });

const ADD_OPERATION: OperationInfo = {
  name: 'add',
  arity: 2,
  symbol: '+',
  description: 'Adds the two operands.',
};

function errorBody(overrides: Partial<ApiErrorBody['error']> = {}): ApiErrorBody {
  return {
    error: {
      code: 'DIVISION_BY_ZERO',
      message: 'Cannot divide by zero.',
      requestId: 'req-body',
      ...overrides,
    },
  };
}

async function rejectionOf(call: Promise<unknown>): Promise<unknown> {
  try {
    await call;
  } catch (error) {
    return error;
  }
  throw new Error('Expected the call to reject');
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.resetModules();
});

describe('calculate', () => {
  it('posts the request as JSON and returns the response', async () => {
    let seenBody: unknown;
    let seenContentType: string | null = null;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        seenBody = await request.json();
        seenContentType = request.headers.get('Content-Type');
        return HttpResponse.json({ operation: 'divide', operands: ['10', '4'], result: '2.5' });
      }),
    );

    const response = await client.calculate({ operation: 'divide', operands: ['10', '4'] });

    expect(response).toEqual({ operation: 'divide', operands: ['10', '4'], result: '2.5' });
    expect(seenBody).toEqual({ operation: 'divide', operands: ['10', '4'] });
    expect(seenContentType).toBe('application/json');
  });

  it('throws ApiError with code, message, status, request ID and fields for a 422', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.json(
          errorBody({
            code: 'INVALID_OPERAND',
            message: 'An operand is not a valid number.',
            fields: { operand_index: 1, reason: 'exponent' },
          }),
          { status: 422 },
        ),
      ),
    );

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '1e5'] }));

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      code: 'INVALID_OPERAND',
      message: 'An operand is not a valid number.',
      status: 422,
      requestId: 'req-body',
      fields: { operand_index: 1, reason: 'exponent' },
    });
  });

  it.each([
    [400, 'BAD_REQUEST'],
    [413, 'PAYLOAD_TOO_LARGE'],
    [415, 'UNSUPPORTED_MEDIA_TYPE'],
    [429, 'RATE_LIMITED'],
    [500, 'INTERNAL'],
    [503, 'SERVICE_UNAVAILABLE'],
  ] as const)('throws ApiError for HTTP %i with code %s', async (status, code) => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.json(errorBody({ code }), { status })));

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ code, status });
  });

  it('accepts every documented error code', async () => {
    for (const code of ERROR_CODES) {
      server.use(
        http.post(CALCULATE_URL, () => HttpResponse.json(errorBody({ code }), { status: 422 })),
      );

      const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

      expect(error).toMatchObject({ code });
    }
  });

  it('throws ApiError UNKNOWN with the status for a non-JSON error body', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.text('<html>Bad gateway</html>', {
          status: 502,
          headers: { 'X-Request-ID': 'req-header' },
        }),
      ),
    );

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ code: 'UNKNOWN', status: 502, requestId: 'req-header' });
    expect((error as ApiError).message).not.toContain('Bad gateway');
  });

  it('throws ApiError UNKNOWN for JSON that is not the error shape', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.json({ oops: true }, { status: 500 })));

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toMatchObject({ code: 'UNKNOWN', status: 500, requestId: undefined });
  });

  it('keeps message and request ID but reports UNKNOWN for a code it does not know', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.json(
          { error: { code: 'BRAND_NEW', message: 'New failure.', requestId: 'req-new' } },
          { status: 422 },
        ),
      ),
    );

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toMatchObject({ code: 'UNKNOWN', message: 'New failure.', requestId: 'req-new' });
  });

  it('throws ApiError INVALID_RESPONSE when a 200 body has the wrong shape', async () => {
    server.use(
      http.post(CALCULATE_URL, () =>
        HttpResponse.json({ operation: 'add', operands: ['1', '2'], result: 3 }),
      ),
    );

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ code: 'INVALID_RESPONSE', status: 200 });
  });

  it('throws ApiError INVALID_RESPONSE when a 200 body is not JSON', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.text('not json')));

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toMatchObject({ code: 'INVALID_RESPONSE' });
  });

  it('throws NetworkError when the connection fails', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.error()));

    const error = await rejectionOf(client.calculate({ operation: 'add', operands: ['1', '2'] }));

    expect(error).toBeInstanceOf(NetworkError);
    expect((error as NetworkError).cause).toBeDefined();
  });

  it('throws TimeoutError when the server is slower than the time limit', async () => {
    server.use(
      http.post(CALCULATE_URL, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );

    const error = await rejectionOf(
      client.calculate({ operation: 'add', operands: ['1', '2'] }, { timeoutMs: 30 }),
    );

    expect(error).toBeInstanceOf(TimeoutError);
    expect((error as TimeoutError).timeoutMs).toBe(30);
  });

  it('uses the client-level time limit by default', async () => {
    server.use(
      http.post(CALCULATE_URL, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );
    const impatient = createApiClient({ baseUrl: BASE_URL, timeoutMs: 20 });

    const error = await rejectionOf(
      impatient.calculate({ operation: 'add', operands: ['1', '2'] }),
    );

    expect(error).toMatchObject({ name: 'TimeoutError', timeoutMs: 20 });
  });

  it('throws CanceledError, not TimeoutError, when the caller aborts mid-request', async () => {
    server.use(
      http.post(CALCULATE_URL, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );
    const controller = new AbortController();

    const pending = rejectionOf(
      client.calculate({ operation: 'add', operands: ['1', '2'] }, { signal: controller.signal }),
    );
    controller.abort();
    const error = await pending;

    expect(error).toBeInstanceOf(CanceledError);
    expect(error).not.toBeInstanceOf(TimeoutError);
  });

  it('throws CanceledError without sending anything when the signal is already aborted', async () => {
    let requests = 0;
    server.use(
      http.post(CALCULATE_URL, () => {
        requests += 1;
        return HttpResponse.json({});
      }),
    );

    const error = await rejectionOf(
      client.calculate({ operation: 'add', operands: ['1', '2'] }, { signal: AbortSignal.abort() }),
    );

    expect(error).toBeInstanceOf(CanceledError);
    expect(requests).toBe(0);
  });

  it('rejects an invalid per-call time limit', async () => {
    const error = await rejectionOf(
      client.calculate({ operation: 'add', operands: ['1', '2'] }, { timeoutMs: 0 }),
    );

    expect(error).toBeInstanceOf(RangeError);
  });
});

describe('createApiClient', () => {
  it('rejects an invalid default time limit', () => {
    expect(() => createApiClient({ timeoutMs: Number.NaN })).toThrow(RangeError);
  });

  it('ignores a trailing slash on the base URL', async () => {
    server.use(http.get(OPERATIONS_URL, () => HttpResponse.json({ operations: [ADD_OPERATION] })));
    const slashed = createApiClient({ baseUrl: `${BASE_URL}/` });

    await expect(slashed.getOperations()).resolves.toEqual({ operations: [ADD_OPERATION] });
  });
});

describe('getOperations', () => {
  it('returns the operations list', async () => {
    let method = '';
    server.use(
      http.get(OPERATIONS_URL, ({ request }) => {
        method = request.method;
        return HttpResponse.json({ operations: [ADD_OPERATION] });
      }),
    );

    const response = await client.getOperations();

    expect(response).toEqual({ operations: [ADD_OPERATION] });
    expect(response.limits).toBeUndefined();
    expect(method).toBe('GET');
  });

  it('returns the server limits when it sends them', async () => {
    const limits = { maxOperandLength: 32, maxExponent: 500, decimalPrecision: 20 };
    server.use(
      http.get(OPERATIONS_URL, () => HttpResponse.json({ operations: [ADD_OPERATION], limits })),
    );

    const response = await client.getOperations();

    expect(response.limits).toEqual(limits);
  });

  it.each([
    ['a non-object', 'big'],
    ['a zero length', { maxOperandLength: 0, maxExponent: 5, decimalPrecision: 5 }],
    ['a fractional length', { maxOperandLength: 1.5, maxExponent: 5, decimalPrecision: 5 }],
    ['a string exponent', { maxOperandLength: 64, maxExponent: '5', decimalPrecision: 5 }],
    ['a missing precision', { maxOperandLength: 64, maxExponent: 5 }],
  ])('rejects limits with %s as an INVALID_RESPONSE', async (_name, limits) => {
    server.use(
      http.get(OPERATIONS_URL, () => HttpResponse.json({ operations: [ADD_OPERATION], limits })),
    );

    const error = await rejectionOf(client.getOperations());

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ code: 'INVALID_RESPONSE' });
  });

  it('accepts an operation this client does not know yet', async () => {
    const modulo = { name: 'modulo', arity: 2, symbol: 'mod', description: 'Remainder.' };
    server.use(
      http.get(OPERATIONS_URL, () => HttpResponse.json({ operations: [ADD_OPERATION, modulo] })),
    );

    const response = await client.getOperations();

    expect(response.operations).toEqual([ADD_OPERATION, modulo]);
  });

  it.each([
    ['an empty operation name', { operations: [{ ...ADD_OPERATION, name: '' }] }],
    ['a non-string operation name', { operations: [{ ...ADD_OPERATION, name: 7 }] }],
    ['a string arity', { operations: [{ ...ADD_OPERATION, arity: '2' }] }],
    ['a null entry', { operations: [null] }],
    [
      'a malformed entry next to an unknown name',
      { operations: [{ ...ADD_OPERATION, name: 'modulo' }, { name: 'add' }] },
    ],
    ['a missing description', { operations: [{ name: 'add', arity: 2, symbol: '+' }] }],
    ['a fractional arity', { operations: [{ ...ADD_OPERATION, arity: 1.5 }] }],
    ['a bare array', [ADD_OPERATION]],
  ])('throws INVALID_RESPONSE for %s', async (_label, body) => {
    server.use(http.get(OPERATIONS_URL, () => HttpResponse.json(body)));

    const error = await rejectionOf(client.getOperations());

    expect(error).toMatchObject({ code: 'INVALID_RESPONSE' });
  });

  it('throws ApiError for a server error', async () => {
    server.use(
      http.get(OPERATIONS_URL, () =>
        HttpResponse.json(errorBody({ code: 'INTERNAL', message: 'Something went wrong.' }), {
          status: 500,
        }),
      ),
    );

    const error = await rejectionOf(client.getOperations());

    expect(error).toMatchObject({ code: 'INTERNAL', status: 500 });
  });

  it('throws CanceledError when the caller aborts', async () => {
    server.use(
      http.get(OPERATIONS_URL, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );
    const controller = new AbortController();

    const pending = rejectionOf(client.getOperations({ signal: controller.signal }));
    controller.abort();

    expect(await pending).toBeInstanceOf(CanceledError);
  });
});

describe('resolveBaseUrl', () => {
  it.each([
    [undefined, ''],
    ['', ''],
    ['   ', ''],
    ['http://localhost:8080', 'http://localhost:8080'],
    ['https://api.example.com//', 'https://api.example.com'],
  ])('turns %j into %j', (raw, expected) => {
    expect(resolveBaseUrl(raw)).toBe(expected);
  });
});

describe('default client', () => {
  it('reads the base URL from VITE_API_BASE_URL', async () => {
    vi.stubEnv('VITE_API_BASE_URL', 'http://env.test');
    vi.resetModules();
    server.use(
      http.get('http://env.test/api/v1/operations', () =>
        HttpResponse.json({ operations: [ADD_OPERATION] }),
      ),
    );

    const { defaultApiClient } = await import('./client');

    await expect(defaultApiClient.getOperations()).resolves.toEqual({
      operations: [ADD_OPERATION],
    });
  });
});
