import { describe, expect, it } from 'vitest';

type PageSessionModule = {
  createPageSessionStore: () => {
    read: (tabKey: string) => unknown;
    write: (tabKey: string, payload: unknown) => void;
    clear: (tabKey: string) => void;
  };
};

async function loadPageSessionModule() {
  const modulePath = './page-session';
  const module = await import(/* @vite-ignore */ modulePath).catch(() => null);

  expect(module).not.toBeNull();
  return module as PageSessionModule;
}

describe('page-session store', () => {
  it('stores and reads payload by tab key', async () => {
    const { createPageSessionStore } = await loadPageSessionModule();
    const store = createPageSessionStore();

    store.write('article:501', { activeTab: 'logs' });

    expect(store.read('article:501')).toEqual({ activeTab: 'logs' });
    expect(store.read('task:77')).toBeUndefined();
  });

  it('replaces payload for an existing tab key', async () => {
    const { createPageSessionStore } = await loadPageSessionModule();
    const store = createPageSessionStore();

    store.write('article:501', { activeTab: 'hits' });
    store.write('article:501', { activeTab: 'changes' });

    expect(store.read('article:501')).toEqual({ activeTab: 'changes' });
  });

  it('clears payload when a tab closes', async () => {
    const { createPageSessionStore } = await loadPageSessionModule();
    const store = createPageSessionStore();

    store.write('task:77:results', { scrollY: 360 });
    store.clear('task:77:results');

    expect(store.read('task:77:results')).toBeUndefined();
  });
});
