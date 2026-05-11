import { readFileSync } from 'node:fs';
import path from 'node:path';

import { theme } from 'antd';
import { describe, expect, it } from 'vitest';

import defaultSettings from '../../config/defaultSettings';
import { adminAntdTheme, adminVisualTokens } from './admin-theme';

describe('admin visual tokens', () => {
  it('keeps the sider dark and the right workspace light', () => {
    expect(defaultSettings.layout).toBe('side');
    expect(defaultSettings.navTheme).toBe('realDark');
    expect(adminVisualTokens.sidebarBg).toMatch(/^#/);
    expect(adminVisualTokens.headerBg).toBe('#ffffff');
    expect(adminVisualTokens.surfaceBg).toBe('#ffffff');
    expect(adminVisualTokens.contentBg).toMatch(/^#f/i);
    expect(adminAntdTheme.algorithm).toBe(theme.defaultAlgorithm);
  });

  it('defines light shell surfaces and avoids black content tokens', () => {
    const globalLess = readFileSync(path.resolve(__dirname, '../global.less'), 'utf8');
    const appSource = readFileSync(path.resolve(__dirname, '../app.tsx'), 'utf8');

    expect(globalLess).toContain('.admin-header');
    expect(globalLess).toContain('.admin-content');
    expect(globalLess).toContain('.admin-light-surface');
    expect(globalLess).not.toMatch(/admin-(?:content|surface|header|tabs)[^#\\n]*#(?:000|000000|111|141414|18181b)/i);
    expect(appSource).toContain('adminAntdTheme');
    expect(appSource).not.toContain('darkAlgorithm');
  });
});
