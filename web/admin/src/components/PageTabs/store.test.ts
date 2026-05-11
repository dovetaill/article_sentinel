import { describe, expect, it } from 'vitest';

import { WORKSPACE_EMPTY_PATH } from './route-meta';
import { reduceTabs, restoreDefaultTabs } from './store';

describe('reduceTabs', () => {
  it('starts with an empty workspace instead of forcing /inspection/tasks', () => {
    const state = restoreDefaultTabs(29);

    expect(state.tabs).toEqual([]);
    expect(state.activeKey).toBe('');
  });

  it('opens the same logical route only once', () => {
    const opened = reduceTabs(restoreDefaultTabs(29), {
      type: 'open',
      href: '/inspection/tasks',
      orgId: 29
    });
    const again = reduceTabs(opened, {
      type: 'open',
      href: '/inspection/tasks?from=menu',
      orgId: 29
    });

    expect(again.tabs).toHaveLength(1);
    expect(again.activeKey).toBe('/inspection/tasks');
  });

  it('falls back to the left tab and then to the empty workspace when closing the active tab', () => {
    const withThreeTabs = ['/inspection/tasks', '/inspection/results', '/rules/keywords'].reduce(
      (state, href) => reduceTabs(state, { type: 'open', href, orgId: 29 }),
      restoreDefaultTabs(29)
    );

    const closedActive = reduceTabs(withThreeTabs, { type: 'close', key: '/rules/keywords' });
    expect(closedActive.activeKey).toBe('/inspection/results');

    const closedAll = reduceTabs(closedActive, { type: 'closeAll' });
    expect(closedAll.tabs).toEqual([]);
    expect(closedAll.activeKey).toBe('');

    const openedWorkspace = reduceTabs(closedAll, {
      type: 'open',
      href: WORKSPACE_EMPTY_PATH,
      orgId: 29
    });
    expect(openedWorkspace.tabs).toEqual([]);
    expect(openedWorkspace.activeKey).toBe('');
  });

  it('keeps only the requested tab when closing others', () => {
    const withThreeTabs = ['/inspection/tasks', '/inspection/results', '/rules/keywords'].reduce(
      (state, href) => reduceTabs(state, { type: 'open', href, orgId: 29 }),
      restoreDefaultTabs(29)
    );

    const closedOthers = reduceTabs(withThreeTabs, {
      type: 'closeOthers',
      key: '/inspection/results'
    });

    expect(closedOthers.tabs).toHaveLength(1);
    expect(closedOthers.tabs[0].key).toBe('/inspection/results');
    expect(closedOthers.activeKey).toBe('/inspection/results');
  });

  it('keeps detail titles friendly', () => {
    const detail = reduceTabs(restoreDefaultTabs(29), {
      type: 'open',
      href: '/content/articles/9',
      orgId: 29
    });

    expect(detail.tabs[0].title).toBe('文章详情');
    expect(detail.tabs[0].closable).toBe(true);
  });
});
