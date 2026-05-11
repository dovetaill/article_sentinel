import { CloseOutlined, MoreOutlined, ReloadOutlined } from '@ant-design/icons';
import { Dropdown, type MenuProps } from 'antd';

import type { TabState } from './store';

type PageTabsProps = {
  state: TabState;
  onActivate: (key: string) => void;
  onClose: (key: string) => void;
  onCloseOthers: (key: string) => void;
  onCloseAll: () => void;
  onRefresh: (key: string) => void;
};

function buildTabItemClassName(isActive: boolean) {
  return ['admin-page-tabs__item', isActive ? 'admin-page-tabs__item--active' : '']
    .filter(Boolean)
    .join(' ');
}

export default function PageTabs(props: PageTabsProps) {
  const { state, onActivate, onClose, onCloseOthers, onCloseAll, onRefresh } = props;

  const activeTab = state.tabs.find((tab) => tab.key === state.activeKey) ?? state.tabs[0];

  const menuItems: MenuProps['items'] = [
    {
      key: 'refresh',
      label: '刷新当前',
      icon: <ReloadOutlined />,
      disabled: !activeTab
    },
    {
      type: 'divider'
    },
    {
      key: 'closeCurrent',
      label: '关闭当前',
      disabled: !activeTab
    },
    {
      key: 'closeOthers',
      label: '关闭其他',
      disabled: !activeTab
    },
    {
      key: 'closeAll',
      label: '关闭全部',
      disabled: state.tabs.length === 0
    }
  ];

  return (
    <div className="admin-page-tabs admin-light-surface">
      <div className="admin-page-tabs__scroll">
        {state.tabs.map((tab) => {
          const isActive = tab.key === state.activeKey;

          return (
            <div key={tab.key} className={buildTabItemClassName(isActive)}>
              <button
                type="button"
                className="admin-page-tabs__tab"
                onClick={() => onActivate(tab.key)}
              >
                <span className="admin-page-tabs__label">{tab.title}</span>
              </button>
              {tab.closable ? (
                <button
                  type="button"
                  className="admin-page-tabs__close"
                  aria-label={`关闭 ${tab.title}`}
                  onClick={(event) => {
                    event.stopPropagation();
                    onClose(tab.key);
                  }}
                >
                  <CloseOutlined />
                </button>
              ) : null}
            </div>
          );
        })}
      </div>
      <Dropdown
        trigger={['click']}
        menu={{
          items: menuItems,
          onClick: ({ key }) => {
            if (!activeTab && key !== 'closeAll') {
              return;
            }

            if (key === 'refresh' && activeTab) {
              onRefresh(activeTab.key);
              return;
            }

            if (key === 'closeCurrent' && activeTab) {
              onClose(activeTab.key);
              return;
            }

            if (key === 'closeOthers' && activeTab) {
              onCloseOthers(activeTab.key);
              return;
            }

            if (key === 'closeAll') {
              onCloseAll();
            }
          }
        }}
      >
        <button type="button" className="admin-page-tabs__more" aria-label="更多标签操作">
          <MoreOutlined />
        </button>
      </Dropdown>
    </div>
  );
}
