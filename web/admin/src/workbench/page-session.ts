type PageSessionSnapshot = Record<string, unknown>;

type PageSessionEntry = {
  live: unknown;
  persisted?: unknown;
};

type WritePageSessionOptions = {
  persisted?: unknown;
};

type PageSessionStore = {
  read: <T>(tabKey: string) => T | undefined;
  write: <T>(tabKey: string, payload: T, options?: WritePageSessionOptions) => void;
  clear: (tabKey: string) => void;
  restore: (snapshot?: PageSessionSnapshot) => void;
  snapshot: (liveTabKeys?: string[]) => PageSessionSnapshot;
};

function createEntryMap(snapshot: PageSessionSnapshot = {}) {
  return Object.fromEntries(
    Object.entries(snapshot).map(([tabKey, value]) => [tabKey, { live: value, persisted: value }]),
  ) as Record<string, PageSessionEntry>;
}

export function createPageSessionStore(initialSnapshot: PageSessionSnapshot = {}): PageSessionStore {
  let entries = createEntryMap(initialSnapshot);

  return {
    read(tabKey) {
      return entries[tabKey]?.live as unknown;
    },
    write(tabKey, payload, options) {
      const nextEntry: PageSessionEntry = { live: payload };

      if (options && 'persisted' in options) {
        nextEntry.persisted = options.persisted;
      } else {
        nextEntry.persisted = payload;
      }

      entries = {
        ...entries,
        [tabKey]: nextEntry
      };
    },
    clear(tabKey) {
      if (!(tabKey in entries)) {
        return;
      }

      const nextEntries = { ...entries };
      delete nextEntries[tabKey];
      entries = nextEntries;
    },
    restore(snapshot) {
      entries = createEntryMap(snapshot);
    },
    snapshot(liveTabKeys) {
      const allowedKeys = liveTabKeys ? new Set(liveTabKeys) : null;
      const snapshot: PageSessionSnapshot = {};

      Object.entries(entries).forEach(([tabKey, entry]) => {
        if (allowedKeys && !allowedKeys.has(tabKey)) {
          return;
        }

        if (entry.persisted === undefined) {
          return;
        }

        snapshot[tabKey] = entry.persisted;
      });

      return snapshot;
    }
  };
}

const orgSessionStores = new Map<number, PageSessionStore>();

function getOrgSessionStore(orgId: number) {
  const existingStore = orgSessionStores.get(orgId);
  if (existingStore) {
    return existingStore;
  }

  const nextStore = createPageSessionStore();
  orgSessionStores.set(orgId, nextStore);
  return nextStore;
}

export function readPageSession<T>(orgId: number, tabKey: string) {
  return getOrgSessionStore(orgId).read<T>(tabKey);
}

export function writePageSession<T>(
  orgId: number,
  tabKey: string,
  payload: T,
  options?: WritePageSessionOptions,
) {
  getOrgSessionStore(orgId).write(tabKey, payload, options);
}

export function clearPageSession(orgId: number, tabKey: string) {
  getOrgSessionStore(orgId).clear(tabKey);
}

export function restorePersistedPageSessions(orgId: number, snapshot?: PageSessionSnapshot) {
  getOrgSessionStore(orgId).restore(snapshot);
}

export function getPersistedPageSessions(orgId: number, liveTabKeys: string[]) {
  return getOrgSessionStore(orgId).snapshot(liveTabKeys);
}
