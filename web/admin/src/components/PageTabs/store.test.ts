import { describe, expect, it } from 'vitest';

import { reduceTabs, restoreDefaultTabs } from './store';

describe('reduceTabs', () => {
  it('opens /inspection/tasks as the fixed base tab', () => {
    const state = restoreDefaultTabs(29);

    expect(state.tabs[0].pathname).toBe('/inspection/tasks');
    expect(state.tabs[0].closable).toBe(false);
  });

  it('deduplicates detail routes by key', () => {
    const next = reduceTabs(restoreDefaultTabs(29), {
      type: 'open',
      href: '/content/articles/42',
      orgId: 29
    });

    const again = reduceTabs(next, {
      type: 'open',
      href: '/content/articles/42?from=results',
      orgId: 29
    });

    expect(again.tabs).toHaveLength(2);
  });
});
