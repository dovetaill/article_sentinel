import type { PropsWithChildren } from 'react';
import { useMemo, useState } from 'react';

import { useLocation } from 'react-router-dom';

import { appRouteGroups, findRouteMeta } from '../../routes';
import { HeaderBar } from './header-bar';
import { SidebarNav } from './sidebar-nav';

export function AdminShell({ children }: PropsWithChildren) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const location = useLocation();
  const routeMeta = useMemo(() => findRouteMeta(location.pathname), [location.pathname]);

  return (
    <div className={`admin-shell${sidebarCollapsed ? ' is-collapsed' : ''}`}>
      <SidebarNav collapsed={sidebarCollapsed} groups={appRouteGroups} />
      <div className="admin-shell__main">
        <HeaderBar
          pageTitle={routeMeta.title}
          sectionLabel={routeMeta.sectionLabel}
          sidebarCollapsed={sidebarCollapsed}
          onToggleSidebar={() => setSidebarCollapsed((current) => !current)}
        />
        <main className="admin-shell__content">
          <div className="admin-shell__content-frame">{children}</div>
        </main>
      </div>
    </div>
  );
}
