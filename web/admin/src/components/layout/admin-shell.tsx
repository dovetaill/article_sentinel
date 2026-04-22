import type { PropsWithChildren } from 'react';

import { appRoutes, type AppRoute } from '../../routes';
import { SidebarNav } from './sidebar-nav';
import { Topbar } from './topbar';

export interface AdminShellProps extends PropsWithChildren {
  activeRoute: AppRoute;
}

export function AdminShell({ activeRoute, children }: AdminShellProps) {
  return (
    <div className="admin-shell">
      <SidebarNav routes={appRoutes} />
      <div className="admin-shell__main">
        <Topbar activeRoute={activeRoute} />
        <main className="admin-shell__content">{children}</main>
      </div>
    </div>
  );
}
