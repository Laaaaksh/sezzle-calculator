import { describe, expect, it } from 'vitest';
import { darkTokens, lightTokens, systemDarkTokens } from '../test/tokens';

describe('tokens.css', () => {
  it('gives the system dark theme exactly the same values as the explicit dark theme', () => {
    // The two copies exist because CSS cannot share a block between a media query and an
    // attribute selector. If they drift, "System" and "Dark" would look different.
    expect(systemDarkTokens).toEqual(darkTokens);
  });

  it('overrides only tokens the light theme defines', () => {
    for (const name of Object.keys(darkTokens)) {
      expect(lightTokens, name).toHaveProperty(name);
    }
  });

  it('uses the colours from the brief', () => {
    expect(lightTokens).toMatchObject({
      '--color-bg': '#f7f6f3',
      '--color-key': '#ffffff',
      '--color-text': '#111827',
      '--color-border': '#e5e7eb',
      '--color-accent': '#5b3df5',
    });
    expect(darkTokens).toMatchObject({
      '--color-bg': '#0b0f14',
      '--color-key': '#151b23',
      '--color-text': '#e6edf3',
      '--color-accent': '#2dd4a7',
    });
  });

  it('keeps every motion duration under 200ms', () => {
    for (const [name, value] of Object.entries(lightTokens)) {
      if (name.startsWith('--duration-')) {
        expect(Number.parseInt(value, 10), name).toBeLessThan(200);
      }
    }
  });
});
