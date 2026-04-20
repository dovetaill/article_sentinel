import { Card, Col, Row, Space, Tag, Typography } from 'antd';
import { Navigate, Route, Routes } from 'react-router-dom';

const { Paragraph, Text, Title } = Typography;

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
  return appRoutes.find((route) => pathname.startsWith(route.path)) ?? appRoutes[0];
}

function RoutePanel({ route }: { route: AppRoute }) {
  return (
    <Card bordered={false} className="route-panel">
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <div>
          <Tag color={route.accent}>{route.label}</Tag>
          <Title level={2}>{route.title}</Title>
          <Paragraph>{route.description}</Paragraph>
        </div>
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}>
            <Card className="route-card" bordered={false}>
              <Text type="secondary">Ready for the next task batch</Text>
              <Paragraph>
                This shell intentionally keeps each module lightweight so the upcoming CRUD pages can slot in
                without reworking the chrome.
              </Paragraph>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card className="route-card route-card-accent" bordered={false}>
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

export function AppRouteOutlet() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/keywords" replace />} />
      {appRoutes.map((route) => (
        <Route key={route.key} path={route.path} element={<RoutePanel route={route} />} />
      ))}
    </Routes>
  );
}
