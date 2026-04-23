import { describe, expect, it } from 'vitest';

import themeCss from './theme.css?raw';
import layoutCss from './layout.css?raw';

describe('admin visual tokens', () => {
  it('uses a neutral shadcn-like palette instead of the previous blue accent system', () => {
    expect(themeCss).toContain('--admin-bg-base: #fafafa;');
    expect(themeCss).toContain('--admin-accent: #18181b;');
    expect(themeCss).toContain('--admin-accent-soft: rgba(24, 24, 27, 0.06);');
    expect(themeCss).not.toContain('#2f5dba');
    expect(themeCss).not.toContain('#1d4ed8');
    expect(themeCss).not.toContain('#274690');
  });

  it('removes the blue brand gradient and keeps shell emphasis neutral', () => {
    expect(layoutCss).not.toContain('linear-gradient(135deg, #1d4ed8, #274690)');
    expect(layoutCss).toContain('background: #18181b;');
    expect(layoutCss).not.toContain('rgba(47, 93, 186, 0.18)');
  });

  it('neutralizes common Ant Design interaction states that would otherwise stay blue', () => {
    expect(layoutCss).toContain('.section-card .ant-tabs-ink-bar');
    expect(layoutCss).toContain('.ant-checkbox-checked .ant-checkbox-inner');
    expect(layoutCss).toContain('.ant-switch.ant-switch-checked');
    expect(layoutCss).toContain('.ant-input-outlined:focus');
  });

  it('compresses shell rhythm and summary density toward the reference layout', () => {
    expect(layoutCss).toContain('width: 34px;');
    expect(layoutCss).toContain('height: 34px;');
    expect(layoutCss).toContain('.admin-topbar {\n  display: flex;\n  gap: 12px;');
    expect(layoutCss).toContain('.summary-card {\n  padding: 12px 14px;');
  });

  it('tightens tables, tabs, and form controls for the final polish pass', () => {
    expect(layoutCss).toContain('min-height: 34px !important;');
    expect(layoutCss).toContain('.section-card .ant-table-wrapper .ant-table-thead > tr > th {\n  padding: 10px 12px;');
    expect(layoutCss).toContain('.section-card .ant-table-wrapper .ant-table-tbody > tr > td {\n  padding: 10px 12px;');
    expect(layoutCss).toContain('.section-card .ant-tabs-tab {\n  padding: 10px 0;');
  });
});
