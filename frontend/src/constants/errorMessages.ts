import type { ApiErrorCode } from '../api';
import type { CalculatorError, ClientFailureCode } from '../features/calculator/state';

/**
 * The one place error copy lives. api/errors.ts, the API client and the UI all read from here,
 * so the same failure never gets two different sentences.
 */
export const GENERIC_ERROR_MESSAGE = 'Something went wrong.';

/** Used as the CanceledError message. Never shown: a cancelled request is not reported to users. */
export const CANCELED_MESSAGE = 'The request was canceled.';

/**
 * Codes with their own copy. UNKNOWN is left out on purpose: it has no meaning of its own, so it
 * takes the generic message with the request ID. Record<..., string> makes a code added to
 * ERROR_CODES without copy here a compile error.
 */
export const ERROR_MESSAGES: Record<
  Exclude<ApiErrorCode | ClientFailureCode, 'UNKNOWN'>,
  string
> = {
  BAD_REQUEST: "That request wasn't valid. Please try again.",
  VALIDATION_FAILED: 'Something is missing from that calculation.',
  UNSUPPORTED_MEDIA_TYPE: 'The app and the server disagree on the format. Please reload the page.',
  PAYLOAD_TOO_LARGE: 'That calculation is too large to send.',
  NOT_FOUND: 'The calculator service could not be found.',
  METHOD_NOT_ALLOWED: 'The server did not accept that request. Please reload the page.',
  INTERNAL: 'The server had a problem. Please try again in a moment.',
  SERVICE_UNAVAILABLE: 'The calculator is temporarily unavailable. Please try again soon.',
  REQUEST_TIMEOUT: 'The calculation took too long. Please try again.',
  UNKNOWN_OPERATION: "The server doesn't support that operation.",
  INVALID_OPERAND_COUNT: 'That operation needs a different number of values.',
  OPERAND_TOO_LARGE: 'That number is too long.',
  INVALID_OPERAND: "That isn't a valid number.",
  DIVISION_BY_ZERO: "You can't divide by zero.",
  INVALID_EXPONENT: 'The exponent must be a whole number.',
  EXPONENT_TOO_LARGE: 'That exponent is too large.',
  RESULT_TOO_LARGE: 'The result is too large to show.',
  NEGATIVE_SQUARE_ROOT: "You can't take the square root of a negative number.",
  INVALID_RESPONSE: "The server sent a reply this app can't read.",
  NETWORK_ERROR: "Couldn't reach the server. Check your connection and try again.",
  TIMEOUT: 'The server took too long to answer. Please try again.',
  RESULT_TOO_LONG_TO_REUSE: 'This result is too long to use in another calculation.',
};

function hasOwnMessage(code: string): code is keyof typeof ERROR_MESSAGES {
  return Object.hasOwn(ERROR_MESSAGES, code);
}

/**
 * The text to show for an error. A code without its own copy (UNKNOWN, or a code that slipped
 * past the types) gets the generic message plus the request ID, so a user can quote it to support.
 */
export function messageForError({ code, requestId }: CalculatorError): string {
  if (hasOwnMessage(code)) {
    return ERROR_MESSAGES[code];
  }
  return requestId === undefined
    ? GENERIC_ERROR_MESSAGE
    : `${GENERIC_ERROR_MESSAGE} Request ID: ${requestId}`;
}
