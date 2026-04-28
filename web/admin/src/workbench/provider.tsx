import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PropsWithChildren
} from 'react';

import { useLocation, useNavigate } from 'react-router-dom';

import { useOrgContext } from '../context/org-context';
import { normalizeWorkbenchPath } from './registry';
import { readWorkbenchSession, writeWorkbenchSession } from './session';
import {
  activateTab as activateTabAction,
  closeAllTabs as closeAllTabsAction,
  closeOtherTabs as closeOtherTabsAction,
  closeTab as closeTabAction,
  closeTabsToLeft as closeTabsToLeftAction,
  closeTabsToRight as closeTabsToRightAction,
  createInitialWorkbenchState,
  openTab as openTabAction,
  replaceTabTitle as replaceTabTitleAction,
  restoreSession as restoreSessionAction,
  type WorkbenchState,
  workbenchReducer
} from './store';
import type { PersistedWorkbenchSession, WorkbenchTab } from './types';

type WorkbenchContextValue = {
  activeKey: string;
  tabs: WorkbenchTab[];
  openTab: (href: string) => void;
  activateTab: (key: string) => void;
  closeTab: (key: string) => void;
  closeOtherTabs: (key: string) => void;
  closeTabsToLeft: (key: string) => void;
  closeTabsToRight: (key: string) => void;
  closeAllTabs: () => void;
  replaceTabTitle: (key: string, title: string) => void;
};

const DEFAULT_ORG_ID = 29;
const BASE_HREF = '/tasks';

const WorkbenchContext = createContext<WorkbenchContextValue | null>(null);

function buildNormalizedHref(pathname: string, search: string) {
  return `${normalizeWorkbenchPath(pathname)}${search}`;
}

function getHrefFromTab(tab: WorkbenchTab | undefined) {
  if (!tab) {
    return BASE_HREF;
  }

  return `${tab.pathname}${tab.search}`;
}

function mergeLocationWithSession(session: PersistedWorkbenchSession | null, currentHref: string, orgId: number) {
  const baseState = session
    ? workbenchReducer(createInitialWorkbenchState({ orgId }), restoreSessionAction(session))
    : createInitialWorkbenchState({ orgId });

  return workbenchReducer(baseState, openTabAction({ href: currentHref, orgId }));
}

export function WorkbenchProvider({ children }: PropsWithChildren) {
  const { activeOrgId, isLoading } = useOrgContext();
  const navigate = useNavigate();
  const location = useLocation();
  const [state, setState] = useState<WorkbenchState>(() => createInitialWorkbenchState({ orgId: DEFAULT_ORG_ID }));
  const restoredOrgRef = useRef<number | null>(null);
  const desiredHrefRef = useRef<string | null>(null);

  const currentOrgId = activeOrgId ?? DEFAULT_ORG_ID;
  const currentHref = buildNormalizedHref(location.pathname, location.search);
  const rawHref = `${location.pathname}${location.search}`;

  const applyAction = useCallback((action: ReturnType<typeof openTabAction> | ReturnType<typeof activateTabAction> | ReturnType<typeof closeTabAction> | ReturnType<typeof closeOtherTabsAction> | ReturnType<typeof closeTabsToLeftAction> | ReturnType<typeof closeTabsToRightAction> | ReturnType<typeof closeAllTabsAction> | ReturnType<typeof replaceTabTitleAction> | ReturnType<typeof restoreSessionAction>) => {
    setState((previous) => {
      const nextState = workbenchReducer(previous, action);
      const activeTab = nextState.tabs.find((tab) => tab.key === nextState.activeKey);

      desiredHrefRef.current = getHrefFromTab(activeTab);
      return nextState;
    });
  }, []);

  useEffect(() => {
    if (isLoading || activeOrgId === null) {
      return;
    }

    const snapshot = readWorkbenchSession(activeOrgId);

    if (restoredOrgRef.current === null) {
      setState((previous) => {
        if (snapshot) {
          return mergeLocationWithSession(snapshot, currentHref, activeOrgId);
        }

        const baseState = previous.orgId === activeOrgId ? previous : createInitialWorkbenchState({ orgId: activeOrgId });
        return workbenchReducer(baseState, openTabAction({ href: currentHref, orgId: activeOrgId }));
      });
      restoredOrgRef.current = activeOrgId;

      if (rawHref !== currentHref) {
        navigate(currentHref, { replace: true });
      }
      return;
    }

    if (restoredOrgRef.current !== activeOrgId) {
      const nextState = snapshot
        ? workbenchReducer(createInitialWorkbenchState({ orgId: activeOrgId }), restoreSessionAction(snapshot))
        : createInitialWorkbenchState({ orgId: activeOrgId });

      setState(nextState);
      restoredOrgRef.current = activeOrgId;
      desiredHrefRef.current = null;

      const activeTab = nextState.tabs.find((tab) => tab.key === nextState.activeKey);
      const nextHref = getHrefFromTab(activeTab);
      if (rawHref !== nextHref) {
        navigate(nextHref, { replace: true });
      }
    }
  }, [activeOrgId, currentHref, isLoading, navigate, rawHref]);

  useEffect(() => {
    if (isLoading || activeOrgId === null || restoredOrgRef.current !== activeOrgId) {
      return;
    }

    if (rawHref !== currentHref) {
      navigate(currentHref, { replace: true });
      return;
    }

    const activeTab = state.tabs.find((tab) => tab.key === state.activeKey);
    const activeHref = getHrefFromTab(activeTab);
    const desiredHref = desiredHrefRef.current;

    if (desiredHref) {
      if (desiredHref === currentHref) {
        desiredHrefRef.current = null;
        return;
      }

      navigate(desiredHref);
      return;
    }

    if (activeHref === currentHref) {
      return;
    }

    setState((previous) => workbenchReducer(previous, openTabAction({ href: currentHref, orgId: activeOrgId })));
  }, [activeOrgId, currentHref, isLoading, navigate, rawHref, state.activeKey, state.tabs]);

  useEffect(() => {
    if (isLoading || activeOrgId === null || restoredOrgRef.current !== activeOrgId || state.orgId !== activeOrgId) {
      return;
    }

    writeWorkbenchSession({
      orgId: activeOrgId,
      activeKey: state.activeKey,
      tabs: state.tabs
    });
  }, [activeOrgId, isLoading, state]);

  const openTab = useCallback((href: string) => {
    applyAction(openTabAction({ href, orgId: currentOrgId }));
  }, [applyAction, currentOrgId]);

  const activateTab = useCallback((key: string) => {
    applyAction(activateTabAction(key));
  }, [applyAction]);

  const closeTab = useCallback((key: string) => {
    applyAction(closeTabAction(key));
  }, [applyAction]);

  const closeOtherTabs = useCallback((key: string) => {
    applyAction(closeOtherTabsAction(key));
  }, [applyAction]);

  const closeTabsToLeft = useCallback((key: string) => {
    applyAction(closeTabsToLeftAction(key));
  }, [applyAction]);

  const closeTabsToRight = useCallback((key: string) => {
    applyAction(closeTabsToRightAction(key));
  }, [applyAction]);

  const closeAllTabs = useCallback(() => {
    applyAction(closeAllTabsAction());
  }, [applyAction]);

  const replaceTabTitle = useCallback((key: string, title: string) => {
    applyAction(replaceTabTitleAction(key, title));
  }, [applyAction]);

  const value = useMemo<WorkbenchContextValue>(() => ({
    activeKey: state.activeKey,
    tabs: state.tabs,
    openTab,
    activateTab,
    closeTab,
    closeOtherTabs,
    closeTabsToLeft,
    closeTabsToRight,
    closeAllTabs,
    replaceTabTitle
  }), [activateTab, closeAllTabs, closeOtherTabs, closeTab, closeTabsToLeft, closeTabsToRight, openTab, replaceTabTitle, state.activeKey, state.tabs]);

  return <WorkbenchContext.Provider value={value}>{children}</WorkbenchContext.Provider>;
}

export function useWorkbenchContext() {
  const context = useContext(WorkbenchContext);
  if (!context) {
    throw new Error('useWorkbench must be used within WorkbenchProvider');
  }

  return context;
}

export function useOptionalWorkbenchContext() {
  return useContext(WorkbenchContext);
}
