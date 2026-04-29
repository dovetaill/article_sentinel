import { AdminShell } from './components/layout/admin-shell';
import { OrgProvider } from './context/org-context';
import { SessionProvider, useSessionContext } from './context/session-context';
import { AppRouteOutlet } from './routes';
import { WorkbenchProvider } from './workbench/provider';

function AuthenticatedApp() {
  const { isLoading, session } = useSessionContext();

  if (isLoading || !session) {
    return null;
  }

  return (
    <OrgProvider>
      <WorkbenchProvider>
        <AdminShell>
          <AppRouteOutlet />
        </AdminShell>
      </WorkbenchProvider>
    </OrgProvider>
  );
}

export default function App() {
  return (
    <SessionProvider>
      <AuthenticatedApp />
    </SessionProvider>
  );
}
