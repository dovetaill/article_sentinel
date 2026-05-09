export type TabDescriptor = {
  key: string;
  pathname: string;
  search: string;
  title: string;
  closable: boolean;
  menuKey: string;
};

type RouteResolver = {
  pattern: string;
  menuKey: string;
  title: (params: Record<string, string>, searchParams: URLSearchParams) => string;
  key: (params: Record<string, string>, searchParams: URLSearchParams) => string;
};

const BASE_PATH = '/inspection/tasks';

const routeResolvers: RouteResolver[] = [
  {
    pattern: '/inspection/tasks',
    menuKey: BASE_PATH,
    title: () => '检测任务',
    key: () => BASE_PATH
  },
  {
    pattern: '/inspection/tasks/create',
    menuKey: BASE_PATH,
    title: () => '新建任务',
    key: () => '/inspection/tasks/create'
  },
  {
    pattern: '/inspection/tasks/:taskId',
    menuKey: BASE_PATH,
    title: ({ taskId }) => `任务#${taskId ?? ''}`,
    key: ({ taskId }) => `task:${taskId ?? ''}`
  },
  {
    pattern: '/inspection/results',
    menuKey: '/inspection/results',
    title: () => '风险结果',
    key: () => '/inspection/results'
  },
  {
    pattern: '/rules/categories',
    menuKey: '/rules/categories',
    title: () => '规则分类',
    key: () => '/rules/categories'
  },
  {
    pattern: '/rules/keywords',
    menuKey: '/rules/keywords',
    title: () => '规则管理',
    key: () => '/rules/keywords'
  },
  {
    pattern: '/content/articles',
    menuKey: '/content/articles',
    title: () => '文稿中心',
    key: () => '/content/articles'
  },
  {
    pattern: '/content/articles/:articleId/rectify',
    menuKey: '/content/articles',
    title: ({ articleId }) => `整改#${articleId ?? ''}`,
    key: ({ articleId }, searchParams) => {
      const taskId = searchParams.get('task_id');
      const resultId = searchParams.get('result_id');

      if (taskId && resultId) {
        return `article:${articleId ?? ''}:rectify:task:${taskId}:result:${resultId}`;
      }

      return `article:${articleId ?? ''}:rectify`;
    }
  },
  {
    pattern: '/content/articles/:articleId',
    menuKey: '/content/articles',
    title: ({ articleId }) => `文稿#${articleId ?? ''}`,
    key: ({ articleId }) => `article:${articleId ?? ''}`
  },
  {
    pattern: '/audit/logs',
    menuKey: '/audit/logs',
    title: () => '操作日志',
    key: () => '/audit/logs'
  }
];

export const workspaceRouteItems = [
  {
    path: '/inspection',
    name: '巡检业务',
    routes: [
      {
        path: '/inspection/tasks',
        name: '检测任务',
        component: './Inspection/TaskList'
      }
    ]
  }
];

function normalizePathname(pathname: string) {
  if (pathname === '/') {
    return BASE_PATH;
  }

  if (pathname.length > 1 && pathname.endsWith('/')) {
    return pathname.replace(/\/+$/, '');
  }

  return pathname;
}

function parseHref(href: string) {
  const url = new URL(href, 'http://admin.local');
  const pathname = normalizePathname(url.pathname);
  const search = url.search || '';

  return {
    pathname,
    search,
    searchParams: new URLSearchParams(search)
  };
}

function matchPattern(pattern: string, pathname: string) {
  const patternSegments = pattern.split('/').filter(Boolean);
  const pathSegments = pathname.split('/').filter(Boolean);

  if (patternSegments.length !== pathSegments.length) {
    return null;
  }

  const params: Record<string, string> = {};

  for (let index = 0; index < patternSegments.length; index += 1) {
    const patternSegment = patternSegments[index];
    const pathSegment = pathSegments[index];

    if (patternSegment.startsWith(':')) {
      params[patternSegment.slice(1)] = pathSegment;
      continue;
    }

    if (patternSegment !== pathSegment) {
      return null;
    }
  }

  return params;
}

export function resolveTabDescriptor(href: string): TabDescriptor {
  const { pathname, search, searchParams } = parseHref(href);

  for (const resolver of routeResolvers) {
    const params = matchPattern(resolver.pattern, pathname);

    if (!params) {
      continue;
    }

    return {
      key: resolver.key(params, searchParams),
      pathname,
      search,
      title: resolver.title(params, searchParams),
      closable: pathname !== BASE_PATH,
      menuKey: resolver.menuKey
    };
  }

  return {
    key: pathname,
    pathname,
    search,
    title: '检测任务',
    closable: pathname !== BASE_PATH,
    menuKey: BASE_PATH
  };
}

export function resolveMenuKey(pathname: string) {
  return resolveTabDescriptor(pathname).menuKey;
}

export function getBasePath() {
  return BASE_PATH;
}
