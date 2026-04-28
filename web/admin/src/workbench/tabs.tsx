import { useEffect, useRef } from 'react';

import { DownOutlined } from '@ant-design/icons';
import { Button, Dropdown, Tabs, Tooltip, type MenuProps } from 'antd';

import { useWorkbench } from './use-workbench';

export function WorkbenchTabs() {
  const {
    activeKey,
    tabs,
    activateTab,
    closeTab,
    closeOtherTabs,
    closeTabsToLeft,
    closeTabsToRight,
    closeAllTabs
  } = useWorkbench();
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const tablist = containerRef.current?.querySelector<HTMLElement>('[role="tablist"]');
    if (!tablist) {
      return;
    }

    tablist.setAttribute('aria-label', '工作台标签');
  }, [tabs.length]);

  const menuItems: MenuProps['items'] = [
    { key: 'current', label: '关闭当前' },
    { key: 'others', label: '关闭其他' },
    { key: 'left', label: '关闭左侧' },
    { key: 'right', label: '关闭右侧' },
    { key: 'all', label: '关闭全部' }
  ];

  const tabItems = tabs.map((tab) => ({
    key: tab.key,
    closable: tab.closable,
    label: (
      <Tooltip title={tab.title}>
        <span className="admin-workbench-tabs__label">{tab.title}</span>
      </Tooltip>
    ),
    children: null
  }));

  return (
    <div ref={containerRef} className="admin-workbench-tabs">
      <Tabs
        activeKey={activeKey}
        className="admin-workbench-tabs__tabs"
        hideAdd
        items={tabItems}
        onChange={activateTab}
        onEdit={(targetKey, action) => {
          if (action === 'remove' && typeof targetKey === 'string') {
            closeTab(targetKey);
          }
        }}
        tabBarExtraContent={{
          right: (
            <Dropdown
              trigger={['click']}
              menu={{
                items: menuItems,
                onClick: ({ key }) => {
                  switch (key) {
                    case 'current':
                      closeTab(activeKey);
                      break;
                    case 'others':
                      closeOtherTabs(activeKey);
                      break;
                    case 'left':
                      closeTabsToLeft(activeKey);
                      break;
                    case 'right':
                      closeTabsToRight(activeKey);
                      break;
                    case 'all':
                      closeAllTabs();
                      break;
                    default:
                      break;
                  }
                }
              }}
            >
              <Button className="admin-workbench-tabs__actions" type="text" aria-label="标签操作">
                <span>标签操作</span>
                <DownOutlined />
              </Button>
            </Dropdown>
          )
        }}
        type="editable-card"
      />
    </div>
  );
}
