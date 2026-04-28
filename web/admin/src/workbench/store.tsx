import { resolveWorkbenchRoute } from './registry';
import type { PersistedWorkbenchSession, WorkbenchCloseAction, WorkbenchTab } from './types';

export type WorkbenchState = {
  orgId: number;
  tabs: WorkbenchTab[];
  activeKey: string;
};

export type WorkbenchAction =
  | { type: 'openTab'; payload: { href: string; orgId: number } }
  | { type: 'activateTab'; payload: { key: string } }
  | { type: 'closeTab'; payload: { key: string } }
  | { type: 'closeOtherTabs'; payload: { key: string } }
  | { type: 'closeTabsToLeft'; payload: { key: string } }
  | { type: 'closeTabsToRight'; payload: { key: string } }
  | { type: 'closeAllTabs' }
  | { type: 'replaceTabTitle'; payload: { key: string; title: string } }
  | { type: 'restoreSession'; payload: PersistedWorkbenchSession };

const BASE_TAB_PATH = '/tasks';

function createTabFromHref(href: string, orgId: number): WorkbenchTab {
  const descriptor = resolveWorkbenchRoute(href);

  return {
    key: descriptor.key,
    pathname: descriptor.pathname,
    search: descriptor.search,
    title: descriptor.title,
    closable: descriptor.closable,
    keepAlive: descriptor.keepAlive,
    orgId
  };
}

function createBaseTab(orgId: number): WorkbenchTab {
  return createTabFromHref(BASE_TAB_PATH, orgId);
}

function ensureBaseTab(tabs: WorkbenchTab[], orgId: number) {
  const baseTab = tabs.find((tab) => tab.key === BASE_TAB_PATH) ?? createBaseTab(orgId);
  const otherTabs = tabs.filter((tab) => tab.key !== BASE_TAB_PATH);

  return [baseTab, ...otherTabs];
}

function findNextActiveKey(tabs: WorkbenchTab[], closingKey: string) {
  const index = tabs.findIndex((tab) => tab.key === closingKey);
  if (index === -1) {
    return BASE_TAB_PATH;
  }

  return tabs[index + 1]?.key ?? tabs[index - 1]?.key ?? BASE_TAB_PATH;
}

function closeTabsByAction(tabs: WorkbenchTab[], key: string, action: Exclude<WorkbenchCloseAction, 'current' | 'all'>, orgId: number) {
  const currentIndex = tabs.findIndex((tab) => tab.key === key);
  if (currentIndex === -1) {
    return ensureBaseTab(tabs, orgId);
  }

  const filteredTabs = tabs.filter((tab, index) => {
    if (!tab.closable) {
      return true;
    }

    switch (action) {
      case 'others':
        return tab.key === key;
      case 'left':
        return index >= currentIndex;
      case 'right':
        return index <= currentIndex;
      default:
        return true;
    }
  });

  return ensureBaseTab(filteredTabs, orgId);
}

export function createInitialWorkbenchState(input: { orgId: number }): WorkbenchState {
  const baseTab = createBaseTab(input.orgId);

  return {
    orgId: input.orgId,
    tabs: [baseTab],
    activeKey: baseTab.key
  };
}

export function openTab(payload: { href: string; orgId: number }): WorkbenchAction {
  return { type: 'openTab', payload };
}

export function activateTab(key: string): WorkbenchAction {
  return { type: 'activateTab', payload: { key } };
}

export function closeTab(key: string): WorkbenchAction {
  return { type: 'closeTab', payload: { key } };
}

export function closeOtherTabs(key: string): WorkbenchAction {
  return { type: 'closeOtherTabs', payload: { key } };
}

export function closeTabsToLeft(key: string): WorkbenchAction {
  return { type: 'closeTabsToLeft', payload: { key } };
}

export function closeTabsToRight(key: string): WorkbenchAction {
  return { type: 'closeTabsToRight', payload: { key } };
}

export function closeAllTabs(): WorkbenchAction {
  return { type: 'closeAllTabs' };
}

export function replaceTabTitle(key: string, title: string): WorkbenchAction {
  return { type: 'replaceTabTitle', payload: { key, title } };
}

export function restoreSession(session: PersistedWorkbenchSession): WorkbenchAction {
  return { type: 'restoreSession', payload: session };
}

export function workbenchReducer(state: WorkbenchState, action: WorkbenchAction): WorkbenchState {
  switch (action.type) {
    case 'openTab': {
      const nextTab = createTabFromHref(action.payload.href, action.payload.orgId);
      const existingIndex = state.tabs.findIndex((tab) => tab.key === nextTab.key);

      if (existingIndex >= 0) {
        const existingTab = state.tabs[existingIndex];
        const shouldReplaceSearch = nextTab.search.length > 0;
        const mergedTab: WorkbenchTab = {
          ...existingTab,
          pathname: nextTab.pathname,
          search: shouldReplaceSearch ? nextTab.search : existingTab.search,
          title: nextTab.title,
          orgId: action.payload.orgId
        };

        const tabs = state.tabs.map((tab, index) => index === existingIndex ? mergedTab : tab);
        return {
          orgId: action.payload.orgId,
          tabs,
          activeKey: mergedTab.key
        };
      }

      return {
        orgId: action.payload.orgId,
        tabs: ensureBaseTab([...state.tabs, nextTab], action.payload.orgId),
        activeKey: nextTab.key
      };
    }

    case 'activateTab':
      if (!state.tabs.some((tab) => tab.key === action.payload.key)) {
        return state;
      }
      return { ...state, activeKey: action.payload.key };

    case 'closeTab': {
      const currentTab = state.tabs.find((tab) => tab.key === action.payload.key);
      if (!currentTab || !currentTab.closable) {
        return state;
      }

      const remainingTabs = ensureBaseTab(
        state.tabs.filter((tab) => tab.key !== action.payload.key),
        state.orgId
      );

      return {
        ...state,
        tabs: remainingTabs,
        activeKey: state.activeKey === action.payload.key ? findNextActiveKey(state.tabs, action.payload.key) : state.activeKey
      };
    }

    case 'closeOtherTabs': {
      const tabs = closeTabsByAction(state.tabs, action.payload.key, 'others', state.orgId);
      return {
        ...state,
        tabs,
        activeKey: tabs.some((tab) => tab.key === action.payload.key) ? action.payload.key : BASE_TAB_PATH
      };
    }

    case 'closeTabsToLeft': {
      const tabs = closeTabsByAction(state.tabs, action.payload.key, 'left', state.orgId);
      return {
        ...state,
        tabs,
        activeKey: state.activeKey
      };
    }

    case 'closeTabsToRight': {
      const tabs = closeTabsByAction(state.tabs, action.payload.key, 'right', state.orgId);
      return {
        ...state,
        tabs,
        activeKey: state.activeKey
      };
    }

    case 'closeAllTabs': {
      const baseTab = createBaseTab(state.orgId);
      return {
        ...state,
        tabs: [baseTab],
        activeKey: baseTab.key
      };
    }

    case 'replaceTabTitle':
      return {
        ...state,
        tabs: state.tabs.map((tab) => tab.key === action.payload.key ? { ...tab, title: action.payload.title } : tab)
      };

    case 'restoreSession': {
      const tabs = ensureBaseTab(action.payload.tabs, action.payload.orgId);
      return {
        orgId: action.payload.orgId,
        tabs,
        activeKey: tabs.some((tab) => tab.key === action.payload.activeKey) ? action.payload.activeKey : BASE_TAB_PATH
      };
    }

    default:
      return state;
  }
}

export function getClosedTabKeys(previousState: WorkbenchState, nextState: WorkbenchState) {
  const nextKeys = new Set(nextState.tabs.map((tab) => tab.key));

  return previousState.tabs
    .filter((tab) => !nextKeys.has(tab.key))
    .map((tab) => tab.key);
}
