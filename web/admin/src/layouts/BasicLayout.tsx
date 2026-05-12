import { useEffect, useMemo, useState } from 'react';

import {
  DownOutlined,
  UserOutlined
} from '@ant-design/icons';
import ProLayout from '@ant-design/pro-layout';
import { Outlet, useLocation, useModel, useNavigate } from '@umijs/max';
import { Avatar, Button, Dropdown, Space, Typography, type MenuProps } from 'antd';

import defaultSettings from '../../config/defaultSettings';
import type { AppInitialState } from '@/app';
import PageTabs from '@/components/PageTabs';
import {
  WORKSPACE_EMPTY_PATH,
  resolveRouteMeta,
  resolveMenuKey,
  workspaceRouteItems
} from '@/components/PageTabs/route-meta';
import { loadStoredTabs, reduceTabs, restoreDefaultTabs, saveStoredTabs, type TabState } from '@/components/PageTabs/store';
import { logout } from '@/services/auth';

export default function BasicLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { initialState } = useModel('@@initialState') as { initialState?: AppInitialState };
  const currentOrgId = initialState?.currentOrgId ?? 0;
  const currentUser = initialState?.currentUser;

  const [collapsed, setCollapsed] = useState(false);
  const [tabState, setTabState] = useState<TabState>(() => restoreDefaultTabs(currentOrgId));

  useEffect(() => {
    if (!currentUser) {
      navigate('/user/login', { replace: true });
    }
  }, [currentUser, navigate]);

  useEffect(() => {
    setTabState(loadStoredTabs(currentOrgId));
  }, [currentOrgId]);

  useEffect(() => {
    if (!currentOrgId) {
      return;
    }

    setTabState((previousState) =>
      reduceTabs(previousState, {
        type: 'open',
        href: `${location.pathname}${location.search}`,
        orgId: currentOrgId
      })
    );
  }, [currentOrgId, location.pathname, location.search]);

  useEffect(() => {
    if (!tabState.orgId) {
      return;
    }

    saveStoredTabs(tabState);
  }, [tabState]);

  const routeTitle = useMemo(() => resolveRouteMeta(location.pathname).name, [location.pathname]);
  const menuRoute = useMemo(
    () => ({
      path: '/',
      routes: workspaceRouteItems
    }),
    []
  );
  const userMenuItems: MenuProps['items'] = useMemo(
    () => [
      {
        key: 'logout',
        label: '退出登录'
      }
    ],
    []
  );

  const navigateToState = (key: string, state: TabState) => {
    if (!key) {
      navigate(WORKSPACE_EMPTY_PATH);
      return;
    }

    const tab = state.tabs.find((item) => item.key === key);

    if (!tab) {
      navigate(WORKSPACE_EMPTY_PATH);
      return;
    }

    navigate(`${tab.pathname}${tab.search}`);
  };

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      navigate('/user/login', { replace: true });
    }
  };

  return (
    <ProLayout
      className="admin-pro-layout"
      {...defaultSettings}
      collapsed={collapsed}
      onCollapse={setCollapsed}
      route={menuRoute}
      location={{ pathname: resolveMenuKey(location.pathname) }}
      menu={{ defaultOpenAll: true }}
      headerRender={false}
      collapsedButtonRender={false}
      menuItemRender={(item, dom) => {
        const path = item.path;
        if (!path) {
          return dom;
        }

        return (
          <a
            onClick={(event) => {
              event.preventDefault();
              navigate(path);
            }}
          >
            {dom}
          </a>
        );
      }}
    >
      <div className="admin-pro-shell">
        <header className="admin-header admin-light-surface" data-testid="admin-header">
          <div className="admin-header__left">
            <Typography.Title level={5} className="admin-header__title">
              {routeTitle}
            </Typography.Title>
          </div>
          <div className="admin-header__right">
            <Dropdown
              trigger={['click']}
              menu={{
                items: userMenuItems,
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    void handleLogout();
                  }
                }
              }}
            >
              <Button type="text" className="admin-user-menu" aria-label="用户菜单">
                <Space size={8}>
                  <Avatar size="small" icon={<UserOutlined />} />
                  <Typography.Text>{currentUser?.nickname ?? currentUser?.orgname ?? '管理员'}</Typography.Text>
                  <DownOutlined />
                </Space>
              </Button>
            </Dropdown>
          </div>
        </header>
        <PageTabs
          state={tabState}
          onActivate={(key) => {
            const nextState = reduceTabs(tabState, { type: 'activate', key });
            setTabState(nextState);
            navigateToState(key, nextState);
          }}
          onClose={(key) => {
            const nextState = reduceTabs(tabState, { type: 'close', key });
            setTabState(nextState);

            if (nextState.activeKey !== tabState.activeKey || key === tabState.activeKey) {
              navigateToState(nextState.activeKey, nextState);
            }
          }}
          onCloseOthers={(key) => {
            const nextState = reduceTabs(tabState, { type: 'closeOthers', key });
            setTabState(nextState);
            navigateToState(nextState.activeKey, nextState);
          }}
          onCloseAll={() => {
            const nextState = reduceTabs(tabState, { type: 'closeAll' });
            setTabState(nextState);
            navigateToState(nextState.activeKey, nextState);
          }}
          onRefresh={(key) => {
            setTabState((previousState) => reduceTabs(previousState, { type: 'refresh', key }));
            navigateToState(key, tabState);
          }}
        />
        <main className="admin-content">
          <Outlet />
        </main>
      </div>
    </ProLayout>
  );
}
