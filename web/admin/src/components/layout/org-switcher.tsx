import { ApartmentOutlined, DownOutlined } from '@ant-design/icons';
import { Button, Dropdown, type MenuProps } from 'antd';

import { useOrgContext } from '../../context/org-context';

export function OrgSwitcher() {
  const { activeOrgId, activeOrgName, isLoading, orgs, setActiveOrgId } = useOrgContext();

  const menu: MenuProps = {
    selectable: true,
    selectedKeys: activeOrgId ? [String(activeOrgId)] : [],
    items: orgs.map((org) => ({
      key: String(org.id),
      label: org.name,
      disabled: !org.enabled
    })),
    onClick: ({ key }) => {
      setActiveOrgId(Number(key));
    }
  };

  return (
    <Dropdown menu={menu} disabled={orgs.length === 0} trigger={['click']}>
      <Button className="admin-header__control" loading={isLoading && !activeOrgName}>
        <ApartmentOutlined />
        <span>{activeOrgName || '一县一端'}</span>
        <DownOutlined />
      </Button>
    </Dropdown>
  );
}
