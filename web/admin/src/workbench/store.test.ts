import { describe, expect, it } from 'vitest';

import {
  closeAllTabs,
  closeOtherTabs,
  createInitialWorkbenchState,
  openTab,
  workbenchReducer
} from './store';

describe('workbench store', () => {
  it('deduplicates single-instance list tabs', () => {
    let state = createInitialWorkbenchState({ orgId: 29 });

    state = workbenchReducer(state, openTab({ href: '/articles', orgId: 29 }));
    state = workbenchReducer(state, openTab({ href: '/articles', orgId: 29 }));

    expect(state.tabs.filter((tab) => tab.key === '/articles')).toHaveLength(1);
  });

  it('replaces list-tab search when reopening with an explicit query', () => {
    let state = createInitialWorkbenchState({ orgId: 29 });

    state = workbenchReducer(state, openTab({ href: '/articles?title=旧条件', orgId: 29 }));
    state = workbenchReducer(state, openTab({ href: '/articles?title=新条件', orgId: 29 }));

    expect(state.tabs.find((tab) => tab.key === '/articles')).toMatchObject({
      key: '/articles',
      search: '?title=新条件'
    });
  });

  it('keeps the base tab while opening multi-instance detail pages', () => {
    let state = createInitialWorkbenchState({ orgId: 29 });

    state = workbenchReducer(state, openTab({ href: '/articles/123', orgId: 29 }));
    state = workbenchReducer(state, openTab({ href: '/articles/456', orgId: 29 }));

    expect(state.tabs.map((tab) => tab.key)).toEqual(['/tasks', 'article:123', 'article:456']);
  });

  it('falls back to the base tab after close-all', () => {
    let state = createInitialWorkbenchState({ orgId: 29 });

    state = workbenchReducer(state, openTab({ href: '/articles/123', orgId: 29 }));
    state = workbenchReducer(state, closeAllTabs());

    expect(state.activeKey).toBe('/tasks');
  });

  it('keeps only the active tab after close-other', () => {
    let state = createInitialWorkbenchState({ orgId: 29 });

    state = workbenchReducer(state, openTab({ href: '/articles', orgId: 29 }));
    state = workbenchReducer(state, openTab({ href: '/tasks/new', orgId: 29 }));
    state = workbenchReducer(state, closeOtherTabs('/tasks/new'));

    expect(state.activeKey).toBe('/tasks/new');
    expect(state.tabs.map((tab) => tab.key)).toEqual(['/tasks/new']);
  });
});
