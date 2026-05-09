import { useEffect, useMemo, useState } from 'react';

import ProLayout from '@ant-design/pro-layout';
import { Outlet, useLocation, useModel, useNavigate } from '@umijs/max';

import defaultSettings from '../../config/defaultSettings';
import PageTabs from '@/components/PageTabs';
import { resolveMenuKey, workspaceRouteItems } from '@/components/PageTabs/route-meta';
import { loadStoredTabs, reduceTabs, restoreDefaultTabs, saveStoredTabs, type TabState } from '@/components/PageTabs/store';
import type { AppInitialState } from '@/app';

export default function BasicLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { initialState } = useModel('@@initialState') as { initialState?: AppInitialState };
  const currentOrgId = initialState?.currentOrgId ?? 0;

  const [tabState, setTabState] = useState<TabState>(() => restoreDefaultTabs(currentOrgId));

  useEffect(() => {
    if (!initialState?.currentUser) {
      navigate('/user/login', { replace: true });
    }
  }, [initialState?.currentUser, navigate]);

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

  const menuRoute = useMemo(
    () => ({
      path: '/',
      routes: workspaceRouteItems
    }),
    []
  );

  const navigateToKey = (key: string, state: TabState) => {
    const tab = state.tabs.find((item) => item.key === key);
    if (!tab) {
      return;
    }

    navigate(`${tab.pathname}${tab.search}`);
  };

  return (
    <ProLayout
      {...defaultSettings}
      route={menuRoute}
      location={{ pathname: resolveMenuKey(location.pathname) }}
      menu={{ defaultOpenAll: true }}
      menuItemRender={(item, dom) => {
        if (!item.path) {
          return dom;
        }

        return (
          <a
            onClick={(event) => {
              event.preventDefault();
              navigate(item.path);
            }}
          >
            {dom}
          </a>
        );
      }}
    >
      <div className="admin-workspace-shell">
        <PageTabs
          state={tabState}
          onActivate={(key) => {
            setTabState((previousState) => reduceTabs(previousState, { type: 'activate', key }));
            navigateToKey(key, tabState);
          }}
          onClose={(key) => {
            const nextState = reduceTabs(tabState, { type: 'close', key });
            setTabState(nextState);

            if (nextState.activeKey !== tabState.activeKey || key === tabState.activeKey) {
              navigateToKey(nextState.activeKey, nextState);
            }
          }}
          onCloseOthers={(key) => {
            const nextState = reduceTabs(tabState, { type: 'closeOthers', key });
            setTabState(nextState);
            navigateToKey(nextState.activeKey, nextState);
          }}
          onCloseAll={() => {
            const nextState = reduceTabs(tabState, { type: 'closeAll' });
            setTabState(nextState);
            navigateToKey(nextState.activeKey, nextState);
          }}
          onRefresh={(key) => {
            setTabState((previousState) => reduceTabs(previousState, { type: 'refresh', key }));
            navigateToKey(key, tabState);
          }}
        />
        <div className="admin-workspace-body">
          <Outlet />
        </div>
      </div>
    </ProLayout>
  );
}
