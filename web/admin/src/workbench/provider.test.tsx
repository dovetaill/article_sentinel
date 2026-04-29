import { ConfigProvider } from 'antd';
import { act, render, screen, waitFor } from '@testing-library/react';
import { useEffect } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../context/org-context';
import { SessionProvider } from '../context/session-context';
import { getWorkbenchSessionKey } from './session';
import { WorkbenchProvider } from './provider';
import { useWorkbench } from './use-workbench';

const { mockedGetSession, mockedLogout } = vi.hoisted(() => ({
  mockedGetSession: vi.fn(),
  mockedLogout: vi.fn(),
}));

vi.mock('../services/auth', () => ({
  getSession: mockedGetSession,
  logout: mockedLogout,
}));

function WorkbenchProbe() {
  const { activeKey, tabs } = useWorkbench();

  return (
    <pre data-testid="workbench-state">
      {JSON.stringify({
        activeKey,
        tabs: tabs.map((tab) => ({
          key: tab.key,
          pathname: tab.pathname,
          search: tab.search,
        })),
      })}
    </pre>
  );
}

function DeferredRestoreObserver() {
  const { activeKey } = useWorkbench();

  useEffect(() => {
    document.body.dataset.activeKey = activeKey;
  }, [activeKey]);

  return null;
}

function renderWorkbench(initialEntries: string[]) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <SessionProvider>
          <OrgProvider>
            <WorkbenchProvider>
              <WorkbenchProbe />
            </WorkbenchProvider>
          </OrgProvider>
        </SessionProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('WorkbenchProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.sessionStorage.clear();
    delete document.body.dataset.activeKey;
    mockedLogout.mockResolvedValue(undefined);
  });

  it('restores tabs from the current session org snapshot and merges the current URL', async () => {
    mockedGetSession.mockResolvedValue({
      id: 90525,
      orgid: 30,
      orgname: '机构30',
      platform: 'chuangqi',
      priv: 'super',
      roleid: '1',
      nickname: '用户A',
      avatar: 'https://example.com/a.png',
    });
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
            orgId: 30,
          },
          {
            key: '/logs',
            pathname: '/logs',
            search: '',
            title: '操作日志',
            closable: true,
            keepAlive: false,
            orgId: 30,
          },
        ],
      }),
    );

    renderWorkbench(['/tasks/new']);

    await waitFor(() => {
      expect(screen.getByTestId('workbench-state')).toHaveTextContent('/tasks/new');
    });
    expect(screen.getByTestId('workbench-state')).toHaveTextContent('/logs');
  });

  it('persists the workbench session under the current session org id', async () => {
    mockedGetSession.mockResolvedValue({
      id: 90525,
      orgid: 30,
      orgname: '机构30',
      platform: 'chuangqi',
      priv: 'super',
      roleid: '1',
      nickname: '用户A',
      avatar: 'https://example.com/a.png',
    });

    renderWorkbench(['/tasks']);

    await waitFor(() => {
      expect(window.sessionStorage.getItem(getWorkbenchSessionKey(30))).not.toBeNull();
    });
    expect(window.sessionStorage.getItem(getWorkbenchSessionKey(29))).toBeNull();
  });

  it('waits for session initialization before restoring a snapshot', async () => {
    let resolveSession:
      | ((value: {
          id: number;
          orgid: number;
          orgname: string;
          platform: string;
          priv: string;
          roleid: string;
          nickname: string;
          avatar?: string;
        }) => void)
      | undefined;
    mockedGetSession.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveSession = resolve;
        }),
    );
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
            orgId: 29,
          },
          {
            key: '/articles',
            pathname: '/articles',
            search: '',
            title: '文稿中心',
            closable: true,
            keepAlive: false,
            orgId: 29,
          },
        ],
      }),
    );

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks']}>
          <SessionProvider>
            <OrgProvider>
              <WorkbenchProvider>
                <DeferredRestoreObserver />
              </WorkbenchProvider>
            </OrgProvider>
          </SessionProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(document.body.dataset.activeKey).not.toBe('/articles');

    await act(async () => {
      resolveSession?.({
        id: 90525,
        orgid: 29,
        orgname: '一县一端',
        platform: 'chuangqi',
        priv: 'super',
        roleid: '1',
        nickname: '用户A',
        avatar: 'https://example.com/a.png',
      });
    });

    await waitFor(() => {
      expect(document.body.dataset.activeKey).toBe('/tasks');
    });
  });
});
