import { AdminShell } from './components/layout/admin-shell';
import { AppRouteOutlet } from './routes';

export default function App() {
  return (
    <AdminShell>
      <AppRouteOutlet />
    </AdminShell>
  );
}
