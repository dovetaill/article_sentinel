import { ConfigProvider } from 'antd';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useEffect } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider, useOrgContext } from '../context/org-context';
import { listOrgs } from '../services/orgs';
import { getWorkbenchSessionKey } from './session';
import { WorkbenchProvider } from './provider';
import { useWorkbench } from './use-workbench';

vi.mock('../services/orgs', () => ({
  listOrgs: vi.fn()
}));

const mockedListOrgs = vi.mocked(listOrgs);

function WorkbenchProbe() {
  const { activeKey, tabs } = useWorkbench();

  return (
    <pre data-testid="workbench-state">
      {JSON.stringify({
        activeKey,
        tabs: tabs.map((tab) => ({
          key: tab.key,
          pathname: tab.pathname,
          search: tab.search
        }))
      })}
    </pre>
  );
}

function SwitchOrgButton() {
  const { setActiveOrgId } = useOrgContext();

  return (
    <button type="button" onClick={() => setActiveOrgId(30)}>
      切换到机构30
    </button>
  );
}

function DeferredRestoreObserver() {
  const { activeKey } = useWorkbench();

  useEffect(() => {
    document.body.dataset.activeKey = activeKey;
  }, [activeKey]);

  return null;
}

describe('WorkbenchProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.sessionStorage.clear();
    delete document.body.dataset.activeKey;
  });

  it('restores tabs from sessionStorage and uses the current URL as the active tab while merging the org snapshot', async () => {
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 }
    ]);
    window.sessionStorage.setItem(
      getWorkbenchSessionKey(29),
      JSON.stringify({
        orgId: 29,
        activeKey: '/articles',
        tabs: [
          {
            key: '/tasks',
            pathname: '/tasks',
            search: '',
            title: '检测任务',
            closable: false,
            keepAlive: false,
            orgId: 29
          },
          {
            key: '/articles',
            pathname: '/articles',
            search: '',
            title: '文稿中心',
            closable: true,
            keepAlive: false,
            orgId: 29
          }
        ]
      }),
    );

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks/new']}>
          <OrgProvider>
            <WorkbenchProvider>
              <WorkbenchProbe />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('workbench-state')).toHaveTextContent('/tasks/new');
    });
    expect(screen.getByTestId('workbench-state')).toHaveTextContent('/articles');
  });

  it('falls back to the base tab for orgs without a snapshot and restores the matching snapshot after an org switch', async () => {
    const user = userEvent.setup();

    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 },
      { id: 30, name: '机构30', cate_id: 0, enabled: true, sort: 2 }
    ]);
    window.sessionStorage.setItem(
      getWorkbenchSessionKey(30),
      JSON.stringify({
        orgId: 30,
        activeKey: '/logs',
        tabs: [
          {
            key: '/tasks',
            pathname: '/tasks',
            search: '',
            title: '检测任务',
            closable: false,
            keepAlive: false,
            orgId: 30
          },
          {
            key: '/logs',
            pathname: '/logs',
            search: '',
            title: '操作日志',
            closable: true,
            keepAlive: false,
            orgId: 30
          }
        ]
      }),
    );

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks']}>
          <OrgProvider>
            <WorkbenchProvider>
              <SwitchOrgButton />
              <WorkbenchProbe />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('workbench-state')).toHaveTextContent('/tasks');
    });
    expect(screen.getByTestId('workbench-state')).not.toHaveTextContent('/logs');

    await user.click(screen.getByRole('button', { name: '切换到机构30' }));

    await waitFor(() => {
      expect(screen.getByTestId('workbench-state')).toHaveTextContent('/logs');
    });
  });

  it('waits for org initialization before restoring a snapshot', async () => {
    let resolveOrgs: ((value: Array<{ id: number; name: string; cate_id: number; enabled: boolean; sort: number }>) => void) | undefined;
    mockedListOrgs.mockImplementationOnce(() => new Promise((resolve) => {
      resolveOrgs = resolve;
    }));
    window.sessionStorage.setItem(
      getWorkbenchSessionKey(29),
      JSON.stringify({
        orgId: 29,
        activeKey: '/articles',
        tabs: [
          {
            key: '/tasks',
            pathname: '/tasks',
            search: '',
            title: '检测任务',
            closable: false,
            keepAlive: false,
            orgId: 29
          },
          {
            key: '/articles',
            pathname: '/articles',
            search: '',
            title: '文稿中心',
            closable: true,
            keepAlive: false,
            orgId: 29
          }
        ]
      }),
    );

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks']}>
          <OrgProvider>
            <WorkbenchProvider>
              <DeferredRestoreObserver />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(document.body.dataset.activeKey).not.toBe('/articles');

    await act(async () => {
      resolveOrgs?.([{ id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 }]);
    });

    await waitFor(() => {
      expect(document.body.dataset.activeKey).toBe('/tasks');
    });
  });
});
