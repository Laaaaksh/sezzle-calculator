import { describe, expect, it } from 'vitest';
import { ERROR_CODES, OPERATION_TYPES, type ErrorCode, type OperationType } from './types';

// Written out by hand on purpose: adding a code to docs/API.md or removing one from types.ts
// must fail a test until both lists are updated together.
const EXPECTED_ERROR_CODES: Record<ErrorCode, true> = {
  BAD_REQUEST: true,
  VALIDATION_FAILED: true,
  UNSUPPORTED_MEDIA_TYPE: true,
  PAYLOAD_TOO_LARGE: true,
  NOT_FOUND: true,
  METHOD_NOT_ALLOWED: true,
  INTERNAL: true,
  SERVICE_UNAVAILABLE: true,
  REQUEST_TIMEOUT: true,
  RATE_LIMITED: true,
  UNKNOWN_OPERATION: true,
  INVALID_OPERAND_COUNT: true,
  OPERAND_TOO_LARGE: true,
  INVALID_OPERAND: true,
  DIVISION_BY_ZERO: true,
  INVALID_EXPONENT: true,
  EXPONENT_TOO_LARGE: true,
  RESULT_TOO_LARGE: true,
  NEGATIVE_SQUARE_ROOT: true,
};

const EXPECTED_OPERATIONS: Record<OperationType, true> = {
  add: true,
  subtract: true,
  multiply: true,
  divide: true,
  power: true,
  sqrt: true,
  percentage: true,
};

// The contract document lives outside the frontend folder. A glob (unlike a plain import) does
// not fail the whole suite when the file is missing, for example when only frontend/ is checked
// out or built in a container. The check then skips with a message that says why.
const API_DOC_PATH = '../../../docs/API.md';
const docs = import.meta.glob<string>('../../../docs/API.md', {
  query: '?raw',
  import: 'default',
  eager: true,
});
const apiDoc: string | undefined = docs[API_DOC_PATH];

function firstColumnOfTableRows(pattern: RegExp): string[] {
  return [...(apiDoc ?? '').matchAll(pattern)].map((match) => match[1] ?? '');
}

describe.skipIf(apiDoc === undefined)('API types match docs/API.md', () => {
  it('lists the same error codes as the constant', () => {
    expect([...ERROR_CODES].sort()).toEqual(Object.keys(EXPECTED_ERROR_CODES).sort());
  });

  it('lists every error code in the docs error table', () => {
    const documented = firstColumnOfTableRows(/^\| `([A-Z_]+)` \| \d{3} \|/gm);

    expect(documented.length).toBeGreaterThan(0);
    expect([...documented].sort()).toEqual([...ERROR_CODES].sort());
  });

  it('lists the same operations as the constant', () => {
    expect([...OPERATION_TYPES].sort()).toEqual(Object.keys(EXPECTED_OPERATIONS).sort());
  });

  it('lists every operation in the docs operations table', () => {
    const documented = firstColumnOfTableRows(/^\| `([a-z]+)` \| \d \| /gm);

    expect(documented.length).toBeGreaterThan(0);
    expect([...documented].sort()).toEqual([...OPERATION_TYPES].sort());
  });
});

// Listed as skipped in the report when the contract check could not run, never a silent pass.
if (apiDoc === undefined) {
  it.skip('docs/API.md not found next to the frontend, so the API contract check was not run', () => {
    // Nothing to run: this entry only exists to make the skip visible.
  });
}
