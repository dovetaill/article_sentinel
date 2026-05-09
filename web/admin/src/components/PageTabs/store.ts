import { getBasePath, resolveTabDescriptor, type TabDescriptor } from './route-meta';

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

function createBaseTab(): TabDescriptor {
  return resolveTabDescriptor(getBasePath());
}

function ensureBaseTab(tabs: TabDescriptor[]) {
  const basePath = getBasePath();
  const baseTab = tabs.find((tab) => tab.pathname === basePath) ?? createBaseTab();
  const otherTabs = tabs.filter((tab) => tab.pathname !== basePath);

  return [baseTab, ...otherTabs];
}

function findNextActiveKey(tabs: TabDescriptor[], closingKey: string) {
  const currentIndex = tabs.findIndex((tab) => tab.key === closingKey);

  if (currentIndex === -1) {
    return getBasePath();
  }

  return tabs[currentIndex + 1]?.key ?? tabs[currentIndex - 1]?.key ?? getBasePath();
}

export function restoreDefaultTabs(orgId: number): TabState {
  const baseTab = createBaseTab();

  return {
    orgId,
    activeKey: baseTab.key,
    tabs: [baseTab]
  };
}

export function reduceTabs(state: TabState, action: TabAction): TabState {
  switch (action.type) {
    case 'open': {
      const nextState = state.orgId === action.orgId ? state : restoreDefaultTabs(action.orgId);
      const nextTab = resolveTabDescriptor(action.href);
      const existingIndex = nextState.tabs.findIndex((tab) => tab.key === nextTab.key);

      if (existingIndex >= 0) {
        const existingTab = nextState.tabs[existingIndex];
        const mergedTab: TabDescriptor = {
          ...existingTab,
          pathname: nextTab.pathname,
          search: nextTab.search || existingTab.search,
          title: nextTab.title,
          menuKey: nextTab.menuKey
        };

        if (
          mergedTab.pathname === existingTab.pathname &&
          mergedTab.search === existingTab.search &&
          mergedTab.title === existingTab.title &&
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
        tabs: ensureBaseTab([...nextState.tabs, nextTab]),
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

      const tabs = ensureBaseTab(state.tabs.filter((tab) => tab.key !== action.key));
      const activeKey = state.activeKey === action.key ? findNextActiveKey(state.tabs, action.key) : state.activeKey;

      return {
        ...state,
        tabs,
        activeKey
      };
    }

    case 'closeOthers': {
      const tabs = ensureBaseTab(state.tabs.filter((tab) => tab.key === action.key || !tab.closable));
      const activeKey = tabs.some((tab) => tab.key === action.key) ? action.key : getBasePath();

      return {
        ...state,
        tabs,
        activeKey
      };
    }

    case 'closeAll': {
      const baseTab = createBaseTab();

      return {
        ...state,
        activeKey: baseTab.key,
        tabs: [baseTab]
      };
    }

    case 'refresh':
      return {
        ...state,
        activeKey: action.key
      };

    case 'restore':
      return {
        orgId: action.state.orgId,
        activeKey: action.state.activeKey,
        tabs: ensureBaseTab(action.state.tabs)
      };

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
