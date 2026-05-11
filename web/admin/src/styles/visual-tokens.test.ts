import { readFileSync } from 'node:fs';
import path from 'node:path';

import { theme } from 'antd';
import { describe, expect, it } from 'vitest';

import { adminAntdTheme, adminVisualTokens } from './admin-theme';
import defaultSettings from '../../config/defaultSettings';

function isDarkHex(hex: string) {
  const normalized = hex.replace('#', '');
  const red = Number.parseInt(normalized.slice(0, 2), 16);
  const green = Number.parseInt(normalized.slice(2, 4), 16);
  const blue = Number.parseInt(normalized.slice(4, 6), 16);
  const luminance = 0.2126 * red + 0.7152 * green + 0.0722 * blue;

  return luminance < 128;
}

function isLightHex(hex: string) {
  return !isDarkHex(hex);
}

describe('admin visual tokens', () => {
  it('keeps only the sidebar dark and the workspace surfaces light', () => {
    expect(isDarkHex(adminVisualTokens.sidebarBg)).toBe(true);
    expect(isLightHex(adminVisualTokens.pageBg)).toBe(true);
    expect(isLightHex(adminVisualTokens.contentBg)).toBe(true);
    expect(isLightHex(adminVisualTokens.surfaceBg)).toBe(true);
    expect(isLightHex(adminVisualTokens.cardBg)).toBe(true);
    expect(isLightHex(adminVisualTokens.tableBg)).toBe(true);
    expect(adminVisualTokens.surfaceBg).toBe('#ffffff');
    expect(adminVisualTokens.cardBg).toBe('#ffffff');
    expect(adminVisualTokens.tableBg).toBe('#ffffff');
    expect(adminAntdTheme.algorithm).toBe(theme.defaultAlgorithm);
  });

  it('defines light workspace CSS variables and avoids dark content surfaces', () => {
    const globalLess = readFileSync(path.resolve(__dirname, '../global.less'), 'utf8');
    const appSource = readFileSync(path.resolve(__dirname, '../app.tsx'), 'utf8');

    expect(defaultSettings.navTheme).toBe('light');
    expect(globalLess).toContain(`--admin-sidebar-bg: ${adminVisualTokens.sidebarBg};`);
    expect(globalLess).toContain(`--admin-page-bg: ${adminVisualTokens.pageBg};`);
    expect(globalLess).toContain(`--admin-content-bg: ${adminVisualTokens.contentBg};`);
    expect(globalLess).toContain(`--admin-surface: ${adminVisualTokens.surfaceBg};`);
    expect(globalLess).toContain(`--admin-card-bg: ${adminVisualTokens.cardBg};`);
    expect(globalLess).toContain(`--admin-table-bg: ${adminVisualTokens.tableBg};`);
    expect(globalLess).not.toMatch(/--admin-(?:content|surface|card|table)[^;]*#(?:111|141414|18181b|000|000000)/i);
    expect(appSource).not.toContain('darkAlgorithm');
  });
});
