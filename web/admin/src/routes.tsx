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
};

export const appRoutes: AppRoute[] = [
  {
    key: 'keywords',
    path: '/keywords',
    label: '关键词规则'
  },
  {
    key: 'tasks',
    path: '/tasks',
    label: '检测任务'
  },
  {
    key: 'results',
    path: '/results',
    label: '风险结果'
  },
  {
    key: 'articles',
    path: '/articles',
    label: '文稿列表'
  },
  {
    key: 'logs',
    path: '/logs',
    label: '操作日志'
  }
];

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
        <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
        <Route path="/logs" element={<LogsPage />} />
      </Routes>
    </Suspense>
  );
}
