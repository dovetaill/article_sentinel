import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import App from './App';
import { appRoutes } from './routes';

const { mockedGetSession, mockedLogout } = vi.hoisted(() => ({
  mockedGetSession: vi.fn(),
  mockedLogout: vi.fn(),
}));

vi.mock('./services/auth', () => ({
  getSession: mockedGetSession,
  logout: mockedLogout,
}));

vi.mock('./routes', async () => {
  const actual = await vi.importActual<typeof import('./routes')>('./routes');
  return {
    ...actual,
    AppRouteOutlet: () => <div data-testid="mock-route-outlet">路由出口</div>
  };
});

const FIXED_LOGIN_URL = 'https://appadmin.cq.qiludev.com/cq-admin/index.html';
const originalLocation = window.location;

describe('App shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetSession.mockResolvedValue({
      id: 90525,
      orgid: 29,
      orgname: '一县一端',
      platform: 'chuangqi',
      priv: 'super',
      roleid: '1',
      nickname: '用户A',
      avatar: 'https://example.com/a.png',
    });
    mockedLogout.mockResolvedValue(undefined);
  });

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    });
  });

  it('shows a workbench tab strip with the base 检测任务 tab when entering from root', async () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(await screen.findByRole('navigation', { name: /主导航/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /规则分类/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /规则管理/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /文稿中心/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /风险结果/i })).not.toBeInTheDocument();
    expect(screen.queryByText('内容巡检与处置工作台')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /一县一端/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /用户A/i })).toBeInTheDocument();

    expect(await screen.findByRole('tablist', { name: /工作台标签/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '检测任务' })).toHaveAttribute('aria-selected', 'true');
    expect(screen.queryByRole('tab', { name: '规则管理' })).not.toBeInTheDocument();
  });

  it('loads the auth session before rendering the admin shell', async () => {
    let resolveSession:
      | ((value: Awaited<ReturnType<typeof getSession>>) => void)
      | undefined;
    mockedGetSession.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveSession = resolve;
        }),
    );

    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(mockedGetSession).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('navigation', { name: /主导航/i })).not.toBeInTheDocument();
    expect(screen.queryByTestId('mock-route-outlet')).not.toBeInTheDocument();

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

    expect(await screen.findByRole('navigation', { name: /主导航/i })).toBeInTheDocument();
    expect(screen.getByTestId('mock-route-outlet')).toBeInTheDocument();
  });

  it('redirects to the fixed login page instead of rendering the app when session bootstrap fails', async () => {
    const assignSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        ...originalLocation,
        assign: assignSpy,
      },
    });
    mockedGetSession.mockRejectedValueOnce(new Error('unauthorized'));

    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(assignSpy).toHaveBeenCalledWith(FIXED_LOGIN_URL);
    });
    expect(screen.queryByRole('navigation', { name: /主导航/i })).not.toBeInTheDocument();
    expect(screen.queryByTestId('mock-route-outlet')).not.toBeInTheDocument();
  });

  it('logs out through the auth service and redirects to the fixed login page', async () => {
    const user = userEvent.setup();
    const assignSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        ...originalLocation,
        assign: assignSpy,
      },
    });

    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    await user.click(await screen.findByRole('button', { name: /用户A/i }));
    await user.click(await screen.findByRole('menuitem', { name: '退出登录' }));

    await waitFor(() => {
      expect(mockedLogout).toHaveBeenCalledTimes(1);
    });
    expect(assignSpy).toHaveBeenCalledWith(FIXED_LOGIN_URL);
  });

  it('treats /articles as the article center rather than the aggregated inspect-results view', () => {
    expect(appRoutes.map((route) => route.label)).toEqual([
      '规则分类',
      '规则管理',
      '检测任务',
      '文稿中心',
      '操作日志'
    ]);
    expect(appRoutes.some((route) => route.path === '/results')).toBe(false);
    expect(appRoutes.some((route) => route.path === '/articles')).toBe(true);
  });

  it('reuses an existing list tab instead of creating duplicates from sidebar switches', async () => {
    const user = userEvent.setup();

    render(
      <MemoryRouter initialEntries={['/tasks']}>
        <App />
      </MemoryRouter>,
    );

    await screen.findByRole('tab', { name: '检测任务' });

    await user.click(screen.getByRole('link', { name: '文稿中心' }));
    await user.click(screen.getByRole('link', { name: '检测任务' }));
    await user.click(screen.getByRole('link', { name: '检测任务' }));

    await waitFor(() => {
      expect(screen.getAllByRole('tab', { name: '检测任务' })).toHaveLength(1);
    });
  });
});
