import { resolveRouteMeta, resolveTabDescriptor, type TabDescriptor } from './route-meta';

export type TabState = {
  orgId: number;
  activeKey: string;
  tabs: TabDescriptor[];
};

export type TabAction =
  | { type: 'open'; href: string; orgId: number }
  | { type: 'activate'; key: string }
  | { type: 'close'; key: string }
  | { type: 'closeOthers'; key: string }
  | { type: 'closeAll' }
  | { type: 'refresh'; key: string }
  | { type: 'restore'; state: TabState };

const STORAGE_PREFIX = 'article-sentinel:page-tabs';

function resolveHrefPathname(href: string) {
  return new URL(href, 'http://admin.local').pathname;
}

function resolveRestoredActiveKey(tabs: TabDescriptor[], activeKey: string) {
  if (tabs.some((tab) => tab.key === activeKey)) {
    return activeKey;
  }

  return tabs[0]?.key ?? '';
}

function findNextActiveKey(tabs: TabDescriptor[], closingKey: string) {
  const currentIndex = tabs.findIndex((tab) => tab.key === closingKey);

  if (currentIndex === -1) {
    return '';
  }

  return tabs[currentIndex - 1]?.key ?? tabs[currentIndex + 1]?.key ?? '';
}

function shouldOpenTab(href: string) {
  return resolveRouteMeta(resolveHrefPathname(href)).opensTab !== false;
}

function sanitizeRestoredTabs(tabs: TabDescriptor[]) {
  return tabs.filter((tab) => resolveRouteMeta(tab.pathname).opensTab !== false);
}

export function restoreDefaultTabs(orgId: number): TabState {
  return {
    orgId,
    activeKey: '',
    tabs: []
  };
}

export function reduceTabs(state: TabState, action: TabAction): TabState {
  switch (action.type) {
    case 'open': {
      const nextState = state.orgId === action.orgId ? state : restoreDefaultTabs(action.orgId);

      if (!shouldOpenTab(action.href)) {
        return nextState;
      }

      const nextTab = resolveTabDescriptor(action.href);
      const existingIndex = nextState.tabs.findIndex((tab) => tab.key === nextTab.key);

      if (existingIndex >= 0) {
        const existingTab = nextState.tabs[existingIndex];
        const mergedTab: TabDescriptor = {
          ...existingTab,
          pathname: nextTab.pathname,
          search: nextTab.search || existingTab.search,
          title: nextTab.title,
          closable: nextTab.closable,
          menuKey: nextTab.menuKey
        };

        if (
          mergedTab.pathname === existingTab.pathname &&
          mergedTab.search === existingTab.search &&
          mergedTab.title === existingTab.title &&
          mergedTab.closable === existingTab.closable &&
          nextState.activeKey === mergedTab.key
        ) {
          return nextState;
        }

        return {
          orgId: action.orgId,
          tabs: nextState.tabs.map((tab, index) => (index === existingIndex ? mergedTab : tab)),
          activeKey: mergedTab.key
        };
      }

      return {
        orgId: action.orgId,
        tabs: [...nextState.tabs, nextTab],
        activeKey: nextTab.key
      };
    }

    case 'activate':
      if (!state.tabs.some((tab) => tab.key === action.key) || state.activeKey === action.key) {
        return state;
      }

      return {
        ...state,
        activeKey: action.key
      };

    case 'close': {
      const currentTab = state.tabs.find((tab) => tab.key === action.key);

      if (!currentTab || !currentTab.closable) {
        return state;
      }

      const tabs = state.tabs.filter((tab) => tab.key !== action.key);
      const activeKey = state.activeKey === action.key ? findNextActiveKey(state.tabs, action.key) : state.activeKey;

      return {
        ...state,
        tabs,
        activeKey
      };
    }

    case 'closeOthers': {
      const tabs = state.tabs.filter((tab) => tab.key === action.key);

      return {
        ...state,
        tabs,
        activeKey: tabs[0]?.key ?? ''
      };
    }

    case 'closeAll':
      return {
        ...state,
        activeKey: '',
        tabs: []
      };

    case 'refresh':
      return {
        ...state,
        activeKey: action.key
      };

    case 'restore': {
      const tabs = sanitizeRestoredTabs(action.state.tabs);

      return {
        orgId: action.state.orgId,
        activeKey: resolveRestoredActiveKey(tabs, action.state.activeKey),
        tabs
      };
    }

    default:
      return state;
  }
}

function getStorageKey(orgId: number) {
  return `${STORAGE_PREFIX}:${orgId}`;
}

export function loadStoredTabs(orgId: number) {
  if (typeof window === 'undefined') {
    return restoreDefaultTabs(orgId);
  }

  const rawValue = window.localStorage.getItem(getStorageKey(orgId));

  if (!rawValue) {
    return restoreDefaultTabs(orgId);
  }

  try {
    const parsed = JSON.parse(rawValue) as TabState;

    if (!parsed || !Array.isArray(parsed.tabs) || typeof parsed.activeKey !== 'string') {
      return restoreDefaultTabs(orgId);
    }

    return reduceTabs(restoreDefaultTabs(orgId), {
      type: 'restore',
      state: {
        orgId,
        activeKey: parsed.activeKey,
        tabs: parsed.tabs
      }
    });
  } catch {
    return restoreDefaultTabs(orgId);
  }
}

export function saveStoredTabs(state: TabState) {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(getStorageKey(state.orgId), JSON.stringify(state));
}
