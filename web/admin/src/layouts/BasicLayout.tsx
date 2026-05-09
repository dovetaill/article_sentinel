import { HomeOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Outlet, useLocation } from '@umijs/max';
import { Breadcrumb, Layout, Typography } from 'antd';

const { Header, Content } = Layout;

const titleMap: Record<string, string> = {
  '/inspection/tasks': '检测任务'
};

export default function BasicLayout() {
  const location = useLocation();
  const title = titleMap[location.pathname] ?? '文章哨兵管理台';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', paddingInline: 24 }}>
        <Typography.Title level={4} style={{ margin: 0, color: '#fff' }}>
          文章哨兵管理台
        </Typography.Title>
      </Header>
      <Content style={{ padding: 24 }}>
        <PageContainer
          header={{
            title,
            breadcrumb: {
              items: [
                { title: <HomeOutlined /> },
                { title }
              ]
            }
          }}
        >
          <Outlet />
        </PageContainer>
      </Content>
    </Layout>
  );
}
