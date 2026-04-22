import {
  AlertOutlined,
  FileSearchOutlined,
  OrderedListOutlined,
  SafetyCertificateOutlined
} from '@ant-design/icons';
import type { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';

import type { AppRoute } from '../../routes';

const routeIconMap: Record<string, ReactNode> = {
  keywords: <SafetyCertificateOutlined />,
  tasks: <OrderedListOutlined />,
  results: <AlertOutlined />,
  logs: <FileSearchOutlined />
};

export interface SidebarNavProps {
  routes: AppRoute[];
}

export function SidebarNav({ routes }: SidebarNavProps) {
  return (
    <aside className="admin-sidebar">
      <div className="admin-sidebar__brand">
        <span className="admin-sidebar__brand-badge">RM</span>
        <div>
          <p className="admin-sidebar__eyebrow">融媒内容安全巡检平台</p>
          <h1>后台管理</h1>
        </div>
      </div>

      <div className="admin-sidebar__section">
        <p className="admin-sidebar__section-title">业务导航</p>
        <nav className="admin-sidebar__nav" aria-label="主导航">
          {routes.map((route) => (
            <NavLink
              key={route.key}
              to={route.path}
              className={({ isActive }) => `admin-sidebar__link${isActive ? ' is-active' : ''}`}
            >
              <span className="admin-sidebar__icon">{routeIconMap[route.key]}</span>
              <span>{route.label}</span>
            </NavLink>
          ))}
        </nav>
      </div>

      <div className="admin-sidebar__footer">
        <p>适用于政府融媒体机构内容巡检与规范处置场景。</p>
      </div>
    </aside>
  );
}
