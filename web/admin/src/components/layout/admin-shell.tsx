import type { PropsWithChildren } from 'react';

import { appRoutes } from '../../routes';
import { SidebarNav } from './sidebar-nav';

export function AdminShell({ children }: PropsWithChildren) {
  return (
    <div className="admin-shell">
      <SidebarNav routes={appRoutes} />
      <div className="admin-shell__main">
        <main className="admin-shell__content">{children}</main>
      </div>
    </div>
  );
}
