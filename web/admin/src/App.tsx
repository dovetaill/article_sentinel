import { useLocation } from 'react-router-dom';

import { AdminShell } from './components/layout/admin-shell';
import { AppRouteOutlet, routeForPath } from './routes';

export default function App() {
  const location = useLocation();
  const activeRoute = routeForPath(location.pathname);

  return (
    <AdminShell activeRoute={activeRoute}>
      <AppRouteOutlet />
    </AdminShell>
  );
}
