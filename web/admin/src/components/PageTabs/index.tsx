import { MoreOutlined, ReloadOutlined } from '@ant-design/icons';
import { Button, Dropdown, Tabs, type MenuProps } from 'antd';

import type { TabState } from './store';

type PageTabsProps = {
  state: TabState;
  onActivate: (key: string) => void;
  onClose: (key: string) => void;
  onCloseOthers: (key: string) => void;
  onCloseAll: () => void;
  onRefresh: (key: string) => void;
};

export default function PageTabs(props: PageTabsProps) {
  const { state, onActivate, onClose, onCloseOthers, onCloseAll, onRefresh } = props;

  const activeTab = state.tabs.find((tab) => tab.key === state.activeKey) ?? state.tabs[0];

  const menuItems: MenuProps['items'] = [
    {
      key: 'refresh',
      label: '刷新当前',
      icon: <ReloadOutlined />
    },
    {
      key: 'closeOthers',
      label: '关闭其他'
    },
    {
      key: 'closeAll',
      label: '关闭全部'
    }
  ];

  return (
    <div className="admin-page-tabs">
      <Tabs
        type="editable-card"
        hideAdd
        activeKey={state.activeKey}
        onChange={onActivate}
        onEdit={(targetKey, action) => {
          if (action === 'remove' && typeof targetKey === 'string') {
            onClose(targetKey);
          }
        }}
        tabBarExtraContent={{
          right: (
            <Dropdown
              menu={{
                items: menuItems,
                onClick: ({ key }) => {
                  if (!activeTab) {
                    return;
                  }

                  if (key === 'refresh') {
                    onRefresh(activeTab.key);
                    return;
                  }

                  if (key === 'closeOthers') {
                    onCloseOthers(activeTab.key);
                    return;
                  }

                  if (key === 'closeAll') {
                    onCloseAll();
                  }
                }
              }}
              trigger={['click']}
            >
              <Button type="text" icon={<MoreOutlined />} />
            </Dropdown>
          )
        }}
        items={state.tabs.map((tab) => ({
          key: tab.key,
          label: tab.title,
          closable: tab.closable
        }))}
      />
    </div>
  );
}
