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
        <span className="admin-sidebar__brand-badge">AS</span>
        <div className="admin-sidebar__brand-copy">
          <p className="admin-sidebar__eyebrow">融媒内容安全巡检平台</p>
          <h1>巡检控制台</h1>
        </div>
      </div>

      <div className="admin-sidebar__section">
        <p className="admin-sidebar__section-title">主导航</p>
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
    </aside>
  );
}
