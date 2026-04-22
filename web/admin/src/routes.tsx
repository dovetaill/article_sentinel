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
    label: '关键词规则',
    title: '关键词规则',
    description: '统一维护巡检词库、匹配范围与处置建议，确保规则口径明确、执行一致。',
    accent: 'gold'
  },
  {
    key: 'tasks',
    path: '/tasks',
    label: '检测任务',
    title: '检测任务',
    description: '统一发起巡检批次，掌握执行状态、扫描规模与命中情况。',
    accent: 'orange'
  },
  {
    key: 'results',
    path: '/results',
    label: '风险结果',
    title: '风险结果',
    description: '集中查看命中文稿、风险等级与处置进展，支撑快速研判与规范处置。',
    accent: 'red'
  },
  {
    key: 'logs',
    path: '/logs',
    label: '操作日志',
    title: '操作日志',
    description: '留存任务与处置过程的关键痕迹，保障记录可查、责任可溯。',
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
              <Text type="secondary">规则执行与任务推进一体呈现</Text>
              <Paragraph>
                当前后台将以统一壳层承载各业务模块，后续页面扩展时无需重复改动公共框架。
              </Paragraph>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card className="route-card route-card-accent" variant="borderless">
              <Text strong>值守提示</Text>
              <Paragraph>
                页面导航、分区标题与表格区域将保持清晰秩序，便于在高频巡检场景下稳定使用。
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
