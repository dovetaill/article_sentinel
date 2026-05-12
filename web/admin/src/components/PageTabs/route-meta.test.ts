import { describe, expect, it } from 'vitest';

import {
  WORKSPACE_EMPTY_PATH,
  resolveMenuKey,
  resolveRouteMeta,
  resolveTabDescriptor
} from './route-meta';

describe('route meta', () => {
  it('derives breadcrumb and menu state from pathname', () => {
    const meta = resolveRouteMeta('/inspection/tasks/42');

    expect(meta.group).toBe('巡检业务');
    expect(meta.breadcrumb).toEqual(['首页', '巡检业务', '任务详情']);
    expect(resolveMenuKey('/inspection/tasks/42')).toBe('/inspection/tasks');
  });

  it('uses friendly tab titles for detail pages', () => {
    expect(resolveTabDescriptor('/content/articles/99').title).toBe('文章详情');
    expect(resolveTabDescriptor('/inspection/tasks/42').title).toBe('任务详情');
  });

  it('marks inspection task tabs as closable and result routes as compatibility-only', () => {
    expect(resolveTabDescriptor('/inspection/tasks').closable).toBe(true);
    expect(resolveRouteMeta('/inspection/results').hiddenInMenu).toBe(true);
    expect(resolveRouteMeta('/inspection/results').opensTab).toBe(false);
  });

  it('marks the workspace route as shell-only', () => {
    const meta = resolveRouteMeta(WORKSPACE_EMPTY_PATH);

    expect(meta.hiddenInMenu).toBe(true);
    expect(meta.opensTab).toBe(false);
  });
});
