import { matchPath } from 'react-router-dom';

import type { WorkbenchRouteDescriptor, WorkbenchRouteKind, WorkbenchRoutePolicy } from './types';

type RouteResolver = {
  pattern: string;
  kind: WorkbenchRouteKind;
  policy: WorkbenchRoutePolicy;
  title: (params: Record<string, string | undefined>, searchParams: URLSearchParams) => string;
  key: (params: Record<string, string | undefined>, searchParams: URLSearchParams) => string;
  fallbackPath: string;
  keepAlive: boolean;
};

const routeResolvers: RouteResolver[] = [
  {
    pattern: '/rules/categories',
    kind: 'list',
    policy: 'single',
    title: () => '规则分类',
    key: () => '/rules/categories',
    fallbackPath: '/tasks',
    keepAlive: false
  },
  {
    pattern: '/rules/keywords',
    kind: 'list',
    policy: 'single',
    title: () => '规则管理',
    key: () => '/rules/keywords',
    fallbackPath: '/tasks',
    keepAlive: false
  },
  {
    pattern: '/tasks/new',
    kind: 'page',
    policy: 'single',
    title: () => '新建任务',
    key: () => '/tasks/new',
    fallbackPath: '/tasks',
    keepAlive: true
  },
  {
    pattern: '/tasks/:taskId/results',
    kind: 'detail',
    policy: 'multi',
    title: ({ taskId }) => `任务结果#${taskId ?? ''}`,
    key: ({ taskId }) => `task:${taskId ?? ''}:results`,
    fallbackPath: '/tasks',
    keepAlive: true
  },
  {
    pattern: '/tasks/:taskId',
    kind: 'detail',
    policy: 'multi',
    title: ({ taskId }) => `任务#${taskId ?? ''}`,
    key: ({ taskId }) => `task:${taskId ?? ''}`,
    fallbackPath: '/tasks',
    keepAlive: true
  },
  {
    pattern: '/tasks',
    kind: 'list',
    policy: 'single',
    title: () => '检测任务',
    key: () => '/tasks',
    fallbackPath: '/tasks',
    keepAlive: false
  },
  {
    pattern: '/results',
    kind: 'list',
    policy: 'single',
    title: () => '风险结果',
    key: () => '/results',
    fallbackPath: '/tasks',
    keepAlive: false
  },
  {
    pattern: '/articles/:articleId/rectify',
    kind: 'detail',
    policy: 'multi',
    title: ({ articleId }) => `整改#${articleId ?? ''}`,
    key: ({ articleId }, searchParams) => {
      const taskId = searchParams.get('task_id');
      const resultId = searchParams.get('result_id');

      if (taskId && resultId) {
        return `article:${articleId ?? ''}:rectify:task:${taskId}:result:${resultId}`;
      }

      return `article:${articleId ?? ''}:rectify`;
    },
    fallbackPath: '/articles',
    keepAlive: true
  },
  {
    pattern: '/articles/:articleId',
    kind: 'detail',
    policy: 'multi',
    title: ({ articleId }) => `文稿#${articleId ?? ''}`,
    key: ({ articleId }) => `article:${articleId ?? ''}`,
    fallbackPath: '/articles',
    keepAlive: true
  },
  {
    pattern: '/articles',
    kind: 'list',
    policy: 'single',
    title: () => '文稿中心',
    key: () => '/articles',
    fallbackPath: '/tasks',
    keepAlive: false
  },
  {
    pattern: '/logs',
    kind: 'list',
    policy: 'single',
    title: () => '操作日志',
    key: () => '/logs',
    fallbackPath: '/tasks',
    keepAlive: false
  }
];

function parseWorkbenchUrl(href: string) {
  const baseUrl = 'http://workbench.local';
  const url = new URL(href, baseUrl);
  const searchIndex = href.indexOf('?');
  const hashIndex = href.indexOf('#');
  const search = searchIndex >= 0
    ? href.slice(searchIndex, hashIndex >= 0 ? hashIndex : undefined)
    : '';

  return {
    pathname: normalizeWorkbenchPath(url.pathname),
    search,
    searchParams: new URLSearchParams(search)
  };
}

export function normalizeWorkbenchPath(pathname: string): string {
  const normalized = pathname !== '/' && pathname.endsWith('/') ? pathname.replace(/\/+$/, '') : pathname;

  switch (normalized) {
    case '/':
      return '/tasks';
    case '/rules':
    case '/keywords':
      return '/rules/keywords';
    default:
      return normalized;
  }
}

export function resolveWorkbenchRoute(href: string): WorkbenchRouteDescriptor {
  const { pathname, search, searchParams } = parseWorkbenchUrl(href);

  for (const resolver of routeResolvers) {
    const match = matchPath({ path: resolver.pattern, end: true }, pathname);
    if (!match) {
      continue;
    }

    return {
      kind: resolver.kind,
      policy: resolver.policy,
      key: resolver.key(match.params, searchParams),
      pathname,
      search,
      title: resolver.title(match.params, searchParams),
      reusable: true,
      closable: pathname !== '/tasks',
      keepAlive: resolver.keepAlive,
      fallbackPath: resolver.fallbackPath,
      supportsAsyncTitle: resolver.kind === 'detail'
    };
  }

  return {
    kind: 'page',
    policy: 'single',
    key: pathname,
    pathname,
    search,
    title: '检测任务',
    reusable: true,
    closable: pathname !== '/tasks',
    keepAlive: false,
    fallbackPath: '/tasks',
    supportsAsyncTitle: false
  };
}
