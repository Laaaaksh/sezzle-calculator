import tokensCss from '../styles/tokens.css?raw';

export type TokenMap = Record<string, string>;

/** Pulls `--name: value;` pairs out of a CSS block body. */
function parseDeclarations(body: string): TokenMap {
  const tokens: TokenMap = {};
  for (const match of body.matchAll(/(--[\w-]+)\s*:\s*([^;]+);/g)) {
    const [, name, value] = match;
    if (name !== undefined && value !== undefined) {
      tokens[name] = value.trim();
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

/** The three places colours are defined in tokens.css. */
export const lightTokens: TokenMap = parseDeclarations(blockBody(/:root\s*\{([^}]*)\}/));
export const darkTokens: TokenMap = parseDeclarations(
  blockBody(/:root\[data-theme=['"]dark['"]\]\s*\{([^}]*)\}/),
);
export const systemDarkTokens: TokenMap = parseDeclarations(
  blockBody(
    /prefers-color-scheme:\s*dark\)\s*\{\s*:root:not\(\[data-theme=['"]light['"]\]\)\s*\{([^}]*)\}/,
  ),
);
