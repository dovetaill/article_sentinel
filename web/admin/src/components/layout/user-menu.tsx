import { DownOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Dropdown, type MenuProps } from 'antd';

const menu: MenuProps = {
  items: [
    {
      key: 'profile',
      label: '当前用户'
    },
    {
      type: 'divider'
    },
    {
      key: 'logout',
      danger: true,
      label: '退出登录'
    }
  ]
};

export function UserMenu() {
  return (
    <Dropdown menu={menu} trigger={['click']}>
      <Button className="admin-header__control">
        <UserOutlined />
        <span>当前用户</span>
        <DownOutlined />
      </Button>
    </Dropdown>
  );
}
