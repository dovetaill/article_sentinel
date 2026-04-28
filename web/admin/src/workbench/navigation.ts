import { useCallback, useContext, type MouseEvent as ReactMouseEvent } from 'react';
import { UNSAFE_LocationContext, useNavigate } from 'react-router-dom';

import { useOptionalWorkbenchContext } from './provider';
import { normalizeWorkbenchPath, resolveWorkbenchRoute } from './registry';
import type { WorkbenchTab } from './types';

type WorkbenchQueryValue = string | number | undefined | null;

export type BuildWorkbenchHrefOptions = {
  returnTo?: string;
  taskId?: WorkbenchQueryValue;
  resultId?: WorkbenchQueryValue;
  categoryId?: WorkbenchQueryValue;
  query?: Record<string, WorkbenchQueryValue>;
};

type BackNavigationOptions = {
  returnTo?: string;
  fallbackTo?: string;
};

const WORKBENCH_BASE = 'http://workbench.local';

function shouldHandleWorkbenchClick(event: Pick<MouseEvent, 'button' | 'metaKey' | 'altKey' | 'ctrlKey' | 'shiftKey'>) {
  return event.button === 0 && !event.metaKey && !event.altKey && !event.ctrlKey && !event.shiftKey;
}

function normalizeHref(href: string) {
  const url = new URL(href, WORKBENCH_BASE);
  const pathname = normalizeWorkbenchPath(url.pathname);

  return `${pathname}${url.search}`;
}

function appendQueryValue(searchParams: URLSearchParams, key: string, value: WorkbenchQueryValue) {
  if (value === undefined || value === null || value === '') {
    return;
  }

  searchParams.set(key, String(value));
}

function findPreferredOpenHref(tabs: WorkbenchTab[], candidateHref: string) {
  const candidateRoute = resolveWorkbenchRoute(candidateHref);
  const matchingTab = tabs.find((tab) => tab.key === candidateRoute.key);

  if (!matchingTab) {
    return null;
  }

  return `${matchingTab.pathname}${matchingTab.search}`;
}

export function buildWorkbenchHref(pathname: string, options: BuildWorkbenchHrefOptions = {}) {
  const url = new URL(pathname, WORKBENCH_BASE);
  const searchParams = new URLSearchParams(url.search);

  appendQueryValue(searchParams, 'return_to', options.returnTo);
  appendQueryValue(searchParams, 'task_id', options.taskId);
  appendQueryValue(searchParams, 'result_id', options.resultId);
  appendQueryValue(searchParams, 'category_id', options.categoryId);

  Object.entries(options.query ?? {}).forEach(([key, value]) => {
    appendQueryValue(searchParams, key, value);
  });

  const normalizedPathname = normalizeWorkbenchPath(url.pathname);
  const search = searchParams.toString();

  return `${normalizedPathname}${search ? `?${search}` : ''}`;
}

export function useWorkbenchNavigation() {
  const locationContext = useContext(UNSAFE_LocationContext);
  const navigate = useNavigate();
  const workbench = useOptionalWorkbenchContext();
  const currentHref = normalizeHref(
    `${locationContext?.location.pathname ?? window.location.pathname}${locationContext?.location.search ?? window.location.search}`
  );

  const open = useCallback((href: string) => {
    const normalizedHref = normalizeHref(href);

    if (workbench) {
      workbench.openTab(normalizedHref);
      return;
    }

    navigate(normalizedHref);
  }, [navigate, workbench]);

  const buildHref = useCallback((pathname: string, options?: BuildWorkbenchHrefOptions) => (
    buildWorkbenchHref(pathname, options)
  ), []);

  const openWithOptions = useCallback((pathname: string, options?: BuildWorkbenchHrefOptions) => {
    const href = buildWorkbenchHref(pathname, options);
    open(href);
    return href;
  }, [open]);

  const onLinkClick = useCallback((event: ReactMouseEvent<HTMLElement>, href: string) => {
    if (!shouldHandleWorkbenchClick(event)) {
      return;
    }

    event.preventDefault();
    open(href);
  }, [open]);

  const goBack = useCallback((options: BackNavigationOptions = {}) => {
    const route = resolveWorkbenchRoute(currentHref);
    const fallbackHref = options.fallbackTo ?? route.fallbackPath;
    const preferredSourceHref = options.returnTo
      ? findPreferredOpenHref(workbench?.tabs ?? [], options.returnTo)
      : null;

    open(preferredSourceHref ?? options.returnTo ?? fallbackHref);
  }, [currentHref, open, workbench?.tabs]);

  return {
    buildHref,
    currentHref,
    goBack,
    onLinkClick,
    open,
    openWithOptions
  };
}
