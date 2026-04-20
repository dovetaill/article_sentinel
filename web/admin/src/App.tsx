import { CompassOutlined } from '@ant-design/icons';
import { ProLayout } from '@ant-design/pro-components';
import { Tag } from 'antd';
import { NavLink, useLocation } from 'react-router-dom';

import { AppRouteOutlet, appRoutes, routeForPath } from './routes';

export default function App() {
  const location = useLocation();
  const activeRoute = routeForPath(location.pathname);

  return (
    <div className="app-shell">
      <ProLayout
        title="Article Sentinel"
        logo={<CompassOutlined />}
        layout="mix"
        splitMenus={false}
        navTheme="light"
        fixedHeader
        fixSiderbar
        location={{ pathname: location.pathname }}
        route={{
          routes: appRoutes.map((route) => ({
            path: route.path,
            name: route.label
          }))
        }}
        menuItemRender={(item, dom) => <NavLink to={item.path ?? '/'}>{dom}</NavLink>}
      >
        <div className="shell-hero">
          <div>
            <Tag color="gold">Inspection Console</Tag>
            <p className="shell-brand">Article Sentinel</p>
            <p>{activeRoute.description}</p>
          </div>
          <div className="shell-highlight">
            <span>Phase 1 shell</span>
            <strong>Keyword scan workflow</strong>
          </div>
        </div>
        <div className="shell-stage">
          <AppRouteOutlet />
        </div>
      </ProLayout>
    </div>
  );
}
