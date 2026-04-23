import { AdminShell } from './components/layout/admin-shell';
import { OrgProvider } from './context/org-context';
import { AppRouteOutlet } from './routes';

export default function App() {
  return (
    <OrgProvider>
      <AdminShell>
        <AppRouteOutlet />
      </AdminShell>
    </OrgProvider>
  );
}
