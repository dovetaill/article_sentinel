import { Suspense, lazy } from 'react';

import { Spin } from 'antd';
import { matchPath, Navigate, Route, Routes } from 'react-router-dom';

const CategoriesPage = lazy(() => import('./pages/categories'));
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
};

export type AppRouteGroup = {
  key: string;
  label: string;
  routes: AppRoute[];
};

type RouteMeta = {
  pattern: string;
  sectionLabel: string;
  title: string;
};

export const appRoutes: AppRoute[] = [
  {
    key: 'rules',
    path: '/rules',
    label: '规则中心'
  },
  {
    key: 'tasks',
    path: '/tasks',
    label: '检测任务'
  },
  {
    key: 'articles',
    path: '/articles',
    label: '文稿中心'
  },
  {
    key: 'logs',
    path: '/logs',
    label: '操作日志'
  }
];

export const appRouteGroups: AppRouteGroup[] = [
  {
    key: 'inspection',
    label: '巡检业务',
    routes: appRoutes.slice(0, 3)
  },
  {
    key: 'audit',
    label: '审计留痕',
    routes: appRoutes.slice(3)
  }
];

const routeMeta: RouteMeta[] = [
  {
    pattern: '/rules/categories',
    sectionLabel: '规则中心',
    title: '规则分类'
  },
  {
    pattern: '/rules/keywords',
    sectionLabel: '规则中心',
    title: '关键词规则'
  },
  {
    pattern: '/tasks/new',
    sectionLabel: '检测任务',
    title: '新建任务'
  },
  {
    pattern: '/tasks/:taskId/results',
    sectionLabel: '检测任务',
    title: '任务结果'
  },
  {
    pattern: '/tasks/:taskId',
    sectionLabel: '检测任务',
    title: '任务详情'
  },
  {
    pattern: '/tasks',
    sectionLabel: '检测任务',
    title: '检测任务'
  },
  {
    pattern: '/articles/:articleId/rectify',
    sectionLabel: '文稿中心',
    title: '整改处置'
  },
  {
    pattern: '/articles/:articleId',
    sectionLabel: '文稿中心',
    title: '文稿详情'
  },
  {
    pattern: '/articles',
    sectionLabel: '文稿中心',
    title: '文稿中心'
  },
  {
    pattern: '/logs',
    sectionLabel: '操作日志',
    title: '操作日志'
  },
  {
    pattern: '/results',
    sectionLabel: '检测任务',
    title: '风险结果'
  }
];

const fallbackRouteMeta: RouteMeta = {
  pattern: '/tasks',
  sectionLabel: '检测任务',
  title: '检测任务'
};

export function findRouteMeta(pathname: string): RouteMeta {
  return routeMeta.find((meta) => matchPath({ path: meta.pattern, end: true }, pathname)) ?? fallbackRouteMeta;
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
        <Route path="/" element={<Navigate to="/rules" replace />} />
        <Route path="/rules" element={<Navigate to="/rules/categories" replace />} />
        <Route path="/keywords" element={<Navigate to="/rules/keywords" replace />} />
        <Route path="/rules/categories" element={<CategoriesPage />} />
        <Route path="/rules/keywords" element={<KeywordsPage />} />
        <Route path="/tasks" element={<TasksPage />} />
        <Route path="/tasks/new" element={<NewTaskPage />} />
        <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
        <Route path="/tasks/:taskId/results" element={<ResultsPage />} />
        <Route path="/results" element={<ResultsPage />} />
        <Route path="/articles" element={<ArticlesPage />} />
        <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
        <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
        <Route path="/logs" element={<LogsPage />} />
      </Routes>
    </Suspense>
  );
}
