import tokensCss from '../styles/tokens.css?raw';

export type TokenMap = Record<string, string>;
export type Scheme = 'light' | 'dark';

const LIGHT_DARK_CALL = 'light-dark(';

/** Pulls `--name: value;` pairs out of a CSS block body. Values may span lines. */
function parseDeclarations(body: string): TokenMap {
  const tokens: TokenMap = {};
  for (const match of body.matchAll(/(--[\w-]+)\s*:\s*([^;]+);/g)) {
    const [, name, value] = match;
    if (name !== undefined && value !== undefined) {
      tokens[name] = value.replace(/\s+/g, ' ').trim();
    }
  }
  return tokens;
}

function blockBody(pattern: RegExp): string {
  const body = pattern.exec(tokensCss)?.[1];
  if (body === undefined) {
    throw new Error(`tokens.css has no block matching ${pattern.source}`);
  }
  return body;
}

/** Splits `a, b` at the one comma that is not inside parentheses. */
function splitTopLevelComma(args: string): [string, string] {
  let depth = 0;
  for (let index = 0; index < args.length; index += 1) {
    const char = args[index];
    if (char === '(') {
      depth += 1;
    } else if (char === ')') {
      depth -= 1;
    } else if (char === ',' && depth === 0) {
      return [args.slice(0, index).trim(), args.slice(index + 1).trim()];
    }
  }
  throw new Error(`light-dark() needs two arguments, got "${args}"`);
}

/** Replaces every light-dark(a, b) in a value with the side that applies to `scheme`. */
export function resolveLightDark(value: string, scheme: Scheme): string {
  const start = value.indexOf(LIGHT_DARK_CALL);
  if (start === -1) {
    return value;
  }
  let depth = 1;
  let end = start + LIGHT_DARK_CALL.length;
  while (depth > 0) {
    const char = value[end];
    if (char === undefined) {
      throw new Error(`Unbalanced light-dark() in "${value}"`);
    }
    depth += char === '(' ? 1 : char === ')' ? -1 : 0;
    end += 1;
  }
  const [light, dark] = splitTopLevelComma(value.slice(start + LIGHT_DARK_CALL.length, end - 1));
  const chosen = scheme === 'light' ? light : dark;
  return resolveLightDark(`${value.slice(0, start)}${chosen}${value.slice(end)}`, scheme);
}

/** Every token exactly as written in the :root block, light-dark() calls included. */
export const rawTokens: TokenMap = parseDeclarations(blockBody(/:root\s*\{([^}]*)\}/));

function resolveAll(scheme: Scheme): TokenMap {
  return Object.fromEntries(
    Object.entries(rawTokens).map(([name, value]) => [name, resolveLightDark(value, scheme)]),
  );
}

/** Token values as the light and the dark scheme see them. */
export const lightTokens: TokenMap = resolveAll('light');
export const darkTokens: TokenMap = resolveAll('dark');

/** The theme rules in tokens.css, for tests that check the scheme switch. */
export const tokensSource: string = tokensCss;
