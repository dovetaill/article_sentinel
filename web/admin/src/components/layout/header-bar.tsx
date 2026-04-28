import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import { Button } from 'antd';

import { OrgSwitcher } from './org-switcher';
import { UserMenu } from './user-menu';

export interface HeaderBarProps {
  pageTitle: string;
  sectionLabel: string;
  sidebarCollapsed: boolean;
  onToggleSidebar: () => void;
}

export function HeaderBar({ pageTitle, sectionLabel, sidebarCollapsed, onToggleSidebar }: HeaderBarProps) {
  const showSectionLabel = sectionLabel.trim().length > 0 && sectionLabel !== pageTitle;

  return (
    <header className="admin-header">
      <div className="admin-header__leading">
        <Button
          type="text"
          className="admin-header__trigger"
          icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          aria-label={sidebarCollapsed ? '展开导航' : '收起导航'}
          onClick={onToggleSidebar}
        />
        <div className="admin-header__title-block">
          {showSectionLabel ? <span className="admin-header__eyebrow">{sectionLabel}</span> : null}
          <h1 className="admin-header__title">{pageTitle}</h1>
        </div>
      </div>

      <div className="admin-header__actions">
        <OrgSwitcher />
        <UserMenu />
      </div>
    </header>
  );
}
