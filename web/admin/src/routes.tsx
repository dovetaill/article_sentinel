import { Suspense, lazy } from 'react';

import { Card, Col, Row, Space, Spin, Tag, Typography } from 'antd';
import { Navigate, Route, Routes } from 'react-router-dom';

const { Paragraph, Text, Title } = Typography;

const KeywordsPage = lazy(() => import('./pages/keywords'));
const TasksPage = lazy(() => import('./pages/tasks'));
const NewTaskPage = lazy(() => import('./pages/tasks/new'));
const ResultsPage = lazy(() => import('./pages/results'));
const LogsPage = lazy(() => import('./pages/logs'));
const RectifyPage = lazy(() => import('./pages/articles/rectify'));

export type AppRoute = {
  key: string;
  path: string;
  label: string;
  title: string;
  description: string;
  accent: string;
};

export const appRoutes: AppRoute[] = [
  {
    key: 'keywords',
    path: '/keywords',
    label: 'Keywords',
    title: 'Rule deck',
    description: 'Tune keyword scopes, risk weights, and action hints before each scan wave.',
    accent: 'gold'
  },
  {
    key: 'tasks',
    path: '/tasks',
    label: 'Tasks',
    title: 'Scan launches',
    description: 'Kick off asynchronous inspection batches and monitor their execution rhythm.',
    accent: 'orange'
  },
  {
    key: 'results',
    path: '/results',
    label: 'Results',
    title: 'Hit stream',
    description: 'Review matched articles, compare risk signals, and send the next action fast.',
    accent: 'red'
  },
  {
    key: 'logs',
    path: '/logs',
    label: 'Logs',
    title: 'Audit trail',
    description: 'Trace who changed what, with request IDs and source IPs preserved end to end.',
    accent: 'blue'
  }
];

export function routeForPath(pathname: string): AppRoute {
  if (pathname.startsWith('/articles/')) {
    return appRoutes[2];
  }
  return appRoutes.find((route) => pathname.startsWith(route.path)) ?? appRoutes[0];
}

function RoutePanel({ route }: { route: AppRoute }) {
  return (
    <Card variant="borderless" className="route-panel">
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <div>
          <Tag color={route.accent}>{route.label}</Tag>
          <Title level={2}>{route.title}</Title>
          <Paragraph>{route.description}</Paragraph>
        </div>
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}>
            <Card className="route-card" variant="borderless">
              <Text type="secondary">Ready for the next task batch</Text>
              <Paragraph>
                This shell intentionally keeps each module lightweight so the upcoming CRUD pages can slot in
                without reworking the chrome.
              </Paragraph>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card className="route-card route-card-accent" variant="borderless">
              <Text strong>Operator cues</Text>
              <Paragraph>
                Focused navigation, bold section titles, and space for table-heavy workflows keep the admin UI
                readable under pressure.
              </Paragraph>
            </Card>
          </Col>
        </Row>
      </Space>
    </Card>
  );
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
        <Route path="/results" element={<ResultsPage />} />
        <Route path="/logs" element={<LogsPage />} />
        <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
      </Routes>
    </Suspense>
  );
}
