import { describe, expect, it } from 'vitest';

import { normalizeWorkbenchPath, resolveWorkbenchRoute } from './registry';

describe('workbench registry', () => {
  it('normalizes alias paths to canonical workbench routes', () => {
    expect(normalizeWorkbenchPath('/')).toBe('/tasks');
    expect(normalizeWorkbenchPath('/keywords')).toBe('/rules/keywords');
  });

  it('resolves single-instance list and work pages', () => {
    expect(resolveWorkbenchRoute('/articles')).toMatchObject({
      kind: 'list',
      key: '/articles',
      reusable: true,
      title: '文稿中心'
    });

    expect(resolveWorkbenchRoute('/tasks/new')).toMatchObject({
      kind: 'page',
      key: '/tasks/new',
      reusable: true,
      title: '新建任务'
    });

    expect(resolveWorkbenchRoute('/results')).toMatchObject({
      kind: 'list',
      key: '/results',
      reusable: true,
      title: '风险结果'
    });
  });

  it('resolves multi-instance detail and rectify routes with stable keys', () => {
    expect(resolveWorkbenchRoute('/articles/123')).toMatchObject({
      kind: 'detail',
      key: 'article:123',
      reusable: true,
      title: '文稿#123'
    });

    expect(resolveWorkbenchRoute('/tasks/88/results')).toMatchObject({
      kind: 'detail',
      key: 'task:88:results',
      title: '任务结果#88'
    });

    expect(resolveWorkbenchRoute('/articles/123/rectify?task_id=77&result_id=11')).toMatchObject({
      kind: 'detail',
      key: 'article:123:rectify:task:77:result:11',
      title: '整改#123'
    });
  });
});
