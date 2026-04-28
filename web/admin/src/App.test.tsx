import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import App from './App';
import { appRoutes } from './routes';
import { listOrgs } from './services/orgs';

vi.mock('./services/orgs', () => ({
  listOrgs: vi.fn()
}));

vi.mock('./routes', async () => {
  const actual = await vi.importActual<typeof import('./routes')>('./routes');
  return {
    ...actual,
    AppRouteOutlet: () => <div data-testid="mock-route-outlet">路由出口</div>
  };
});

const mockedListOrgs = vi.mocked(listOrgs);

describe('App shell', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 }
    ]);
  });

  it('shows a workbench tab strip with the base 检测任务 tab when entering from root', async () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole('navigation', { name: /主导航/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /规则分类/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /规则管理/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /文稿中心/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /风险结果/i })).not.toBeInTheDocument();
    expect(screen.queryByText('内容巡检与处置工作台')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /一县一端/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /当前用户|退出登录/i })).toBeInTheDocument();

    expect(await screen.findByRole('tablist', { name: /工作台标签/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '检测任务' })).toHaveAttribute('aria-selected', 'true');
    expect(screen.queryByRole('tab', { name: '规则管理' })).not.toBeInTheDocument();
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
