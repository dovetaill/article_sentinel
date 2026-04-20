import { describe, expect, it } from 'vitest';

import { chunkForModule } from './chunks';

describe('chunkForModule', () => {
  it('splits heavy UI and runtime vendors into stable chunks', () => {
    expect(chunkForModule('/repo/web/admin/node_modules/react/index.js')).toBe('react-vendor');
    expect(chunkForModule('/repo/web/admin/node_modules/react-router-dom/dist/index.js')).toBe('router-vendor');
    expect(chunkForModule('/repo/web/admin/node_modules/antd/es/layout/index.js')).toBe('antd-layout');
    expect(chunkForModule('/repo/web/admin/node_modules/@ant-design/icons/es/index.js')).toBe('icons-vendor');
    expect(chunkForModule('/repo/web/admin/node_modules/rc-field-form/es/index.js')).toBe('rc-vendor');
    expect(chunkForModule('/repo/web/admin/node_modules/@ant-design/pro-components/es/index.js')).toBe('pro-components');
    expect(chunkForModule('/repo/web/admin/node_modules/@ant-design/pro-form/es/index.js')).toBe('pro-form');
    expect(chunkForModule('/repo/web/admin/node_modules/@ant-design/pro-table/es/index.js')).toBe('pro-table');
    expect(chunkForModule('/repo/web/admin/node_modules/@ant-design/pro-layout/es/index.js')).toBe('pro-layout');
    expect(chunkForModule('/repo/web/admin/node_modules/dayjs/dayjs.min.js')).toBe('vendor');
  });

  it('ignores first-party source files', () => {
    expect(chunkForModule('/repo/web/admin/src/App.tsx')).toBeUndefined();
  });
});
