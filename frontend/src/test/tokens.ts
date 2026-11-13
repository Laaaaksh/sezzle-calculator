import postcss, { type Declaration, type Root, type Rule } from 'postcss';
import tokensCss from '../styles/tokens.css?raw';

export type TokenMap = Record<string, string>;
export type Scheme = 'light' | 'dark';

const LIGHT_DARK_CALL = 'light-dark(';

export const tokensRoot: Root = postcss.parse(tokensCss, { from: 'tokens.css' });

export function ruleFor(selector: string): Rule {
  const rule = tokensRoot.nodes.find(
    (node): node is Rule => node.type === 'rule' && node.selector === selector,
  );
  if (rule === undefined) {
    throw new Error(`tokens.css has no rule for ${selector}`);
  }
  return rule;
}

export function declarationsOf(prop: string): Declaration[] {
  const found: Declaration[] = [];
  tokensRoot.walkDecls(prop, (decl) => {
    found.push(decl);
  });
  return found;
}

function customProperties(rule: Rule): TokenMap {
  const tokens: TokenMap = {};
  rule.each((node) => {
    if (node.type === 'decl' && node.prop.startsWith('--')) {
      tokens[node.prop] = node.value.replace(/\s+/g, ' ').trim();
    }
  });
  return tokens;
}

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

export const rawTokens: TokenMap = customProperties(ruleFor(':root'));

function resolveAll(scheme: Scheme): TokenMap {
  return Object.fromEntries(
    Object.entries(rawTokens).map(([name, value]) => [name, resolveLightDark(value, scheme)]),
  );
}

export const lightTokens: TokenMap = resolveAll('light');
export const darkTokens: TokenMap = resolveAll('dark');
