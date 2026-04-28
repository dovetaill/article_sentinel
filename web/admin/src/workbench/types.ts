export type WorkbenchRouteKind = 'list' | 'detail' | 'page';

export type WorkbenchRoutePolicy = 'single' | 'multi';

export type WorkbenchCloseAction = 'current' | 'others' | 'left' | 'right' | 'all';

export type WorkbenchRouteDescriptor = {
  kind: WorkbenchRouteKind;
  policy: WorkbenchRoutePolicy;
  key: string;
  pathname: string;
  search: string;
  title: string;
  reusable: boolean;
  closable: boolean;
  keepAlive: boolean;
  fallbackPath: string;
  supportsAsyncTitle: boolean;
};

export type WorkbenchTab = {
  key: string;
  pathname: string;
  search: string;
  title: string;
  closable: boolean;
  keepAlive: boolean;
  orgId: number;
};

export type PersistedWorkbenchSession = {
  orgId: number;
  activeKey: string;
  tabs: WorkbenchTab[];
  pageSessions?: Record<string, unknown>;
};
