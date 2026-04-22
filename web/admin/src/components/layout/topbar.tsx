import { BellOutlined, ClockCircleOutlined, TeamOutlined } from '@ant-design/icons';

import type { AppRoute } from '../../routes';

export interface TopbarProps {
  activeRoute: AppRoute;
}

export function Topbar({ activeRoute }: TopbarProps) {
  return (
    <header className="admin-topbar">
      <div>
        <p className="admin-topbar__eyebrow">融媒内容安全巡检平台</p>
        <div className="admin-topbar__title-row">
          <h2>{activeRoute.title}</h2>
          <span className="admin-topbar__pill">值守中</span>
        </div>
        <p className="admin-topbar__description">{activeRoute.description}</p>
      </div>

      <div className="admin-topbar__meta">
        <div className="admin-topbar__meta-card">
          <TeamOutlined />
          <div>
            <span>适用机构</span>
            <strong>政府融媒体机构</strong>
          </div>
        </div>
        <div className="admin-topbar__meta-card">
          <ClockCircleOutlined />
          <div>
            <span>巡检时段</span>
            <strong>全天值守</strong>
          </div>
        </div>
        <div className="admin-topbar__meta-card">
          <BellOutlined />
          <div>
            <span>提示状态</span>
            <strong>规则已启用</strong>
          </div>
        </div>
      </div>
    </header>
  );
}
