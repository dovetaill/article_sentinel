export type TabDescriptor = {
  key: string;
  pathname: string;
  search: string;
  title: string;
  closable: boolean;
  menuKey: string;
};

export type WorkspaceRouteMeta = {
  path: string;
  name: string;
  group: string;
  breadcrumb: string[];
  menuKey: string;
  tabTitle?: string | ((params: Record<string, string>, search: URLSearchParams) => string);
  tabKey?: (params: Record<string, string>, search: URLSearchParams) => string;
  closable?: boolean;
  hiddenInMenu?: boolean;
  opensTab?: boolean;
  component?: string;
  routes?: WorkspaceRouteMeta[];
};

export const WORKSPACE_EMPTY_PATH = '/workspace';

const BASE_PATH = '/inspection/tasks';

const workspaceRouteTree: WorkspaceRouteMeta[] = [
  {
    path: WORKSPACE_EMPTY_PATH,
    name: '工作区',
    group: '工作区',
    breadcrumb: ['首页', '工作区'],
    menuKey: BASE_PATH,
    hiddenInMenu: true,
    opensTab: false,
    component: './Workspace/Empty'
  },
  {
    path: '/inspection',
    name: '巡检业务',
    group: '巡检业务',
    breadcrumb: ['首页', '巡检业务'],
    menuKey: BASE_PATH,
    opensTab: false,
    routes: [
      {
        path: '/inspection/tasks',
        name: '检测任务',
        group: '巡检业务',
        breadcrumb: ['首页', '巡检业务', '检测任务'],
        menuKey: BASE_PATH,
        tabTitle: '检测任务',
        closable: true,
        component: './Inspection/TaskList'
      },
      {
        path: '/inspection/tasks/create',
        name: '新建任务',
        group: '巡检业务',
        breadcrumb: ['首页', '巡检业务', '新建任务'],
        menuKey: BASE_PATH,
        tabTitle: '新建任务',
        hiddenInMenu: true,
        component: './Inspection/TaskCreate'
      },
      {
        path: '/inspection/tasks/:taskId',
        name: '任务详情',
        group: '巡检业务',
        breadcrumb: ['首页', '巡检业务', '任务详情'],
        menuKey: BASE_PATH,
        tabTitle: '任务详情',
        tabKey: ({ taskId }) => `task:${taskId ?? ''}`,
        hiddenInMenu: true,
        component: './Inspection/TaskDetail'
      },
      {
        path: '/inspection/results',
        name: '风险结果',
        group: '巡检业务',
        breadcrumb: ['首页', '巡检业务', '风险结果'],
        menuKey: BASE_PATH,
        tabTitle: '风险结果',
        hiddenInMenu: true,
        opensTab: false,
        component: './Inspection/ResultList'
      }
    ]
  },
  {
    path: '/rules',
    name: '规则中心',
    group: '规则中心',
    breadcrumb: ['首页', '规则中心'],
    menuKey: '/rules/keywords',
    opensTab: false,
    routes: [
      {
        path: '/rules/categories',
        name: '规则分类',
        group: '规则中心',
        breadcrumb: ['首页', '规则中心', '规则分类'],
        menuKey: '/rules/categories',
        tabTitle: '规则分类',
        component: './Rules/CategoryList'
      },
      {
        path: '/rules/keywords',
        name: '关键词规则',
        group: '规则中心',
        breadcrumb: ['首页', '规则中心', '关键词规则'],
        menuKey: '/rules/keywords',
        tabTitle: '关键词规则',
        component: './Rules/KeywordList'
      }
    ]
  },
  {
    path: '/content',
    name: '内容中心',
    group: '内容中心',
    breadcrumb: ['首页', '内容中心'],
    menuKey: '/content/articles',
    opensTab: false,
    routes: [
      {
        path: '/content/articles',
        name: '文稿列表',
        group: '内容中心',
        breadcrumb: ['首页', '内容中心', '文稿列表'],
        menuKey: '/content/articles',
        tabTitle: '文稿列表',
        component: './Content/ArticleList'
      },
      {
        path: '/content/articles/:articleId',
        name: '文章详情',
        group: '内容中心',
        breadcrumb: ['首页', '内容中心', '文章详情'],
        menuKey: '/content/articles',
        tabTitle: '文章详情',
        tabKey: ({ articleId }) => `article:${articleId ?? ''}`,
        hiddenInMenu: true,
        component: './Content/ArticleDetail'
      },
      {
        path: '/content/articles/:articleId/rectify',
        name: '内容整改',
        group: '内容中心',
        breadcrumb: ['首页', '内容中心', '内容整改'],
        menuKey: '/content/articles',
        tabTitle: '内容整改',
        tabKey: ({ articleId }, searchParams) => {
          const taskId = searchParams.get('task_id');
          const resultId = searchParams.get('result_id');

          if (taskId && resultId) {
            return `article:${articleId ?? ''}:rectify:task:${taskId}:result:${resultId}`;
          }

          return `article:${articleId ?? ''}:rectify`;
        },
        hiddenInMenu: true,
        component: './Content/ArticleRectify'
      }
    ]
  },
  {
    path: '/audit',
    name: '审计留痕',
    group: '审计留痕',
    breadcrumb: ['首页', '审计留痕'],
    menuKey: '/audit/logs',
    opensTab: false,
    routes: [
      {
        path: '/audit/logs',
        name: '操作日志',
        group: '审计留痕',
        breadcrumb: ['首页', '审计留痕', '操作日志'],
        menuKey: '/audit/logs',
        tabTitle: '操作日志',
        component: './Audit/OperationLogList'
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

function flattenRouteMetas(routes: WorkspaceRouteMeta[]) {
  return routes.flatMap((route) => {
    if (!route.routes?.length) {
      return [route];
    }

    return flattenRouteMetas(route.routes);
  });
}

const workspaceRouteMetas = flattenRouteMetas(workspaceRouteTree);
const fallbackRouteMeta = workspaceRouteMetas.find((route) => route.path === BASE_PATH)!;

type MatchedWorkspaceRouteMeta = {
  meta: WorkspaceRouteMeta;
  params: Record<string, string>;
};

function matchWorkspaceRouteMeta(pathname: string): MatchedWorkspaceRouteMeta {
  const normalizedPathname = normalizePathname(pathname);

  for (const meta of workspaceRouteMetas) {
    const params = matchPattern(meta.path, normalizedPathname);

    if (params) {
      return {
        meta,
        params
      };
    }
  }

  return {
    meta: fallbackRouteMeta,
    params: {}
  };
}

function resolveTabTitle(
  meta: WorkspaceRouteMeta,
  params: Record<string, string>,
  searchParams: URLSearchParams
) {
  if (typeof meta.tabTitle === 'function') {
    return meta.tabTitle(params, searchParams);
  }

  if (meta.tabTitle) {
    return meta.tabTitle;
  }

  return meta.name;
}

export const workspaceRouteItems = workspaceRouteTree.map((route) => {
  if (!route.routes?.length) {
    return {
      path: route.path,
      name: route.name,
      component: route.component,
      hideInMenu: route.hiddenInMenu
    };
  }

  return {
    path: route.path,
    name: route.name,
    routes: route.routes.map((childRoute) => ({
      path: childRoute.path,
      name: childRoute.name,
      component: childRoute.component,
      hideInMenu: childRoute.hiddenInMenu
    }))
  };
});

export function resolveRouteMeta(pathname: string): WorkspaceRouteMeta {
  return matchWorkspaceRouteMeta(pathname).meta;
}

export function resolveBreadcrumb(pathname: string) {
  return resolveRouteMeta(pathname).breadcrumb;
}

export function resolveBreadcrumbItems(pathname: string) {
  return resolveBreadcrumb(pathname).map((title, index) => ({
    key: `${index}-${title}`,
    title
  }));
}

export function resolveTabDescriptor(href: string): TabDescriptor {
  const { pathname, search, searchParams } = parseHref(href);
  const { meta, params } = matchWorkspaceRouteMeta(pathname);

  return {
    key: meta.tabKey?.(params, searchParams) ?? pathname,
    pathname,
    search,
    title: resolveTabTitle(meta, params, searchParams),
    closable: meta.closable ?? pathname !== BASE_PATH,
    menuKey: meta.menuKey
  };
}

export function resolveMenuKey(pathname: string) {
  return resolveRouteMeta(pathname).menuKey;
}

export function getBasePath() {
  return BASE_PATH;
}
