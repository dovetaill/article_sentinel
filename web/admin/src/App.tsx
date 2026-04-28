import { AdminShell } from './components/layout/admin-shell';
import { OrgProvider } from './context/org-context';
import { AppRouteOutlet } from './routes';
import { WorkbenchProvider } from './workbench/provider';

export default function App() {
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
