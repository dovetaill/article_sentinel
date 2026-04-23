import { Suspense, lazy } from 'react';

import { Spin } from 'antd';
import { Navigate, Route, Routes } from 'react-router-dom';

const KeywordsPage = lazy(() => import('./pages/keywords'));
const TasksPage = lazy(() => import('./pages/tasks'));
const NewTaskPage = lazy(() => import('./pages/tasks/new'));
const TaskDetailPage = lazy(() => import('./pages/tasks/detail'));
const ResultsPage = lazy(() => import('./pages/results'));
const LogsPage = lazy(() => import('./pages/logs'));
const ArticlesPage = lazy(() => import('./pages/articles'));
const ArticleDetailPage = lazy(() => import('./pages/articles/detail'));
const RectifyPage = lazy(() => import('./pages/articles/rectify'));

export type AppRoute = {
  key: string;
  path: string;
  label: string;
  title: string;
  description: string;
  accent?: string;
};

export const appRoutes: AppRoute[] = [
  {
    key: 'keywords',
    path: '/keywords',
    label: '关键词规则',
    title: '关键词规则',
    description: '统一维护规则词库、匹配方式与建议处置。',
    accent: 'gold'
  },
  {
    key: 'tasks',
    path: '/tasks',
    label: '检测任务',
    title: '检测任务',
    description: '统一发起巡检批次，查看执行状态与命中概况。',
    accent: 'orange'
  },
  {
    key: 'results',
    path: '/results',
    label: '风险结果',
    title: '风险结果',
    description: '集中查看命中文稿、风险等级与处置进展。',
    accent: 'red'
  },
  {
    key: 'logs',
    path: '/logs',
    label: '操作日志',
    title: '操作日志',
    description: '留存任务执行与稿件处置过程的关键记录。'
  }
];

const hiddenRoutes: AppRoute[] = [
  {
    key: 'task-new',
    path: '/tasks/new',
    label: '新建检测任务',
    title: '新建检测任务',
    description: '设置巡检范围与关键词条件，提交新的检测批次。'
  },
  {
    key: 'task-detail',
    path: '/tasks/:taskId',
    label: '任务详情',
    title: '任务详情',
    description: '集中查看任务配置、命中结果与关联日志。'
  },
  {
    key: 'articles',
    path: '/articles',
    label: '文稿列表',
    title: '文稿列表',
    description: '按稿件维度查看巡检命中、处置状态与最近任务。'
  },
  {
    key: 'article-detail',
    path: '/articles/:articleId',
    label: '文稿详情',
    title: '文稿详情',
    description: '集中查看单篇文稿的命中情况、正文快照与整改入口。'
  },
  {
    key: 'article-rectify',
    path: '/articles/:articleId/rectify',
    label: '内容整改',
    title: '内容整改',
    description: '围绕当前稿件进行标题、摘要与正文修订。'
  }
];

const routeMatchers: Array<{ match: (pathname: string) => boolean; route: AppRoute }> = [
  { match: (pathname) => pathname === '/tasks/new', route: hiddenRoutes[0] },
  { match: (pathname) => pathname.startsWith('/tasks/') && pathname !== '/tasks/new', route: hiddenRoutes[1] },
  { match: (pathname) => pathname === '/articles', route: hiddenRoutes[2] },
  { match: (pathname) => pathname.startsWith('/articles/') && pathname.endsWith('/rectify'), route: hiddenRoutes[4] },
  { match: (pathname) => pathname.startsWith('/articles/'), route: hiddenRoutes[3] },
  ...appRoutes.map((route) => ({
    match: (pathname: string) => pathname.startsWith(route.path),
    route
  }))
];

export function routeForPath(pathname: string): AppRoute {
  return routeMatchers.find((item) => item.match(pathname))?.route ?? appRoutes[0];
}

function RouteFallback() {
  return (
    <div style={{ padding: '48px 0', textAlign: 'center' }}>
      <Spin size="large" />
    </div>
  );
}

export function AppRouteOutlet() {
  return (
    <Suspense fallback={<RouteFallback />}>
      <Routes>
        <Route path="/" element={<Navigate to="/keywords" replace />} />
        <Route path="/keywords" element={<KeywordsPage />} />
        <Route path="/tasks" element={<TasksPage />} />
        <Route path="/tasks/new" element={<NewTaskPage />} />
        <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
        <Route path="/results" element={<ResultsPage />} />
        <Route path="/articles" element={<ArticlesPage />} />
        <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
        <Route path="/logs" element={<LogsPage />} />
        <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
      </Routes>
    </Suspense>
  );
}
