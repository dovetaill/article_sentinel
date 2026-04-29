import { ApartmentOutlined } from '@ant-design/icons';
import { Button } from 'antd';

import { useOrgContext } from '../../context/org-context';

export function OrgSwitcher() {
  const { activeOrgName, isLoading } = useOrgContext();

  return (
    <Button className="admin-header__control" loading={isLoading && !activeOrgName}>
      <ApartmentOutlined />
      <span>{activeOrgName || '一县一端'}</span>
    </Button>
  );
}
