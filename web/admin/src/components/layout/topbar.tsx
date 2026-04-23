import type { AppRoute } from '../../routes';

export interface TopbarProps {
  activeRoute: AppRoute;
}

export function Topbar({ activeRoute }: TopbarProps) {
  return (
    <header className="admin-topbar">
      <div className="admin-topbar__copy">
        <div className="admin-topbar__breadcrumbs" aria-label="页面路径">
          <span>控制台</span>
          <span>/</span>
          <span>{activeRoute.label}</span>
        </div>
        <div className="admin-topbar__title-row">
          <h2>{activeRoute.title}</h2>
        </div>
        <p className="admin-topbar__description">{activeRoute.description}</p>
      </div>
    </header>
  );
}
