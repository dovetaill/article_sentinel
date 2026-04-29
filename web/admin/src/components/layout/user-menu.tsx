import { DownOutlined, UserOutlined } from '@ant-design/icons';
import { Avatar, Button, Dropdown, type MenuProps } from 'antd';

import { useSessionContext } from '../../context/session-context';

export function UserMenu() {
  const { logout, session } = useSessionContext();

  const menu: MenuProps = {
    items: [
      {
        key: 'logout',
        danger: true,
        label: '退出登录'
      }
    ],
    onClick: ({ key }) => {
      if (key === 'logout') {
        void logout();
      }
    }
  };

  return (
    <Dropdown menu={menu} trigger={['click']}>
      <Button className="admin-header__control">
        <Avatar size="small" src={session?.avatar} icon={<UserOutlined />} />
        <span>{session?.nickname || '当前用户'}</span>
        <DownOutlined />
      </Button>
    </Dropdown>
  );
}
