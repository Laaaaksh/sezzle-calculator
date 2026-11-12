import { describe, expect, it } from 'vitest';
import indexHtml from '../index.html?raw';

const document = new DOMParser().parseFromString(indexHtml, 'text/html');

function viewport(): Map<string, string> {
  const content = document.querySelector('meta[name="viewport"]')?.getAttribute('content') ?? '';
  return new Map(
    content.split(',').map((part) => {
      const [key = '', value = ''] = part.split('=');
      return [key.trim().toLowerCase(), value.trim().toLowerCase()];
    }),
  );
}

describe('index.html', () => {
  it('sets the page language for screen readers', () => {
    expect(document.documentElement.getAttribute('lang')).toBe('en');
  });

  it('has a viewport that follows the device width', () => {
    expect(viewport().get('width')).toBe('device-width');
  });

  it('does not block pinch zoom', () => {
    const settings = viewport();
    expect(['no', '0']).not.toContain(settings.get('user-scalable'));
  });

  it('does not cap zoom below 500 percent', () => {
    const max = viewport().get('maximum-scale');
    expect(max === undefined || Number(max) >= 5).toBe(true);
  });

  it('loads every script from a file, so a strict Content-Security-Policy still works', () => {
    const scripts = [...document.querySelectorAll('script')];
    expect(scripts.length).toBeGreaterThan(0);
    for (const script of scripts) {
      expect(script.getAttribute('src')).toBeTruthy();
    }
  });
});
