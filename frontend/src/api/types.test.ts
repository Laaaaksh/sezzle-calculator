import { describe, expect, it } from 'vitest';
import apiDoc from '../../../docs/API.md?raw';
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

function firstColumnOfTableRows(pattern: RegExp): string[] {
  return [...apiDoc.matchAll(pattern)].map((match) => match[1] ?? '');
}

describe('API types match docs/API.md', () => {
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
