import {
  FileSearchOutlined,
  FolderOpenOutlined,
  OrderedListOutlined,
  SafetyCertificateOutlined
} from '@ant-design/icons';
import type { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';

import type { AppRouteGroup } from '../../routes';

const routeIconMap: Record<string, ReactNode> = {
  categories: <SafetyCertificateOutlined />,
  keywords: <SafetyCertificateOutlined />,
  tasks: <OrderedListOutlined />,
  articles: <FolderOpenOutlined />,
  logs: <FileSearchOutlined />
};

export interface SidebarNavProps {
  collapsed: boolean;
  groups: AppRouteGroup[];
}

export function SidebarNav({ collapsed, groups }: SidebarNavProps) {
  return (
    <aside className="admin-sidebar">
      <div className="admin-sidebar__brand">
        <span className="admin-sidebar__brand-badge">AS</span>
        <div className="admin-sidebar__brand-copy">
          <p className="admin-sidebar__brand-title">文章安全巡检后台</p>
        </div>
      </div>

      <nav className="admin-sidebar__nav" aria-label="主导航">
        {groups.map((group) => (
          <div key={group.key} className="admin-sidebar__group">
            <p className="admin-sidebar__section-title">{group.label}</p>
            <div className="admin-sidebar__group-links">
              {group.routes.map((route) => (
                <NavLink
                  key={route.key}
                  to={route.path}
                  className={({ isActive }) => `admin-sidebar__link${isActive ? ' is-active' : ''}`}
                  end={route.path !== '/tasks' && route.path !== '/articles'}
                  aria-label={collapsed ? route.label : undefined}
                >
                  <span className="admin-sidebar__icon">{routeIconMap[route.key]}</span>
                  <span className="admin-sidebar__label">{route.label}</span>
                </NavLink>
              ))}
            </div>
          </div>
        ))}
      </nav>
    </aside>
  );
}
