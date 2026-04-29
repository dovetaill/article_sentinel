import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { WorkbenchProvider } from '../../workbench/provider';
import TasksPage from './index';
import { deleteTask, listTasks } from '../../services/tasks';

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn(),
  deleteTask: vi.fn()
}));

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
}));

const mockedListTasks = vi.mocked(listTasks);
const mockedDeleteTask = vi.mocked(deleteTask);

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/tasks']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <TasksPage />
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function readLocationState() {
  const text = screen.getByTestId('location-probe').textContent ?? '';
  const [pathname, search = ''] = text.split('?');

  return {
    pathname,
    searchParams: new URLSearchParams(search)
  };
}

describe('TasksPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListTasks.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 2,
      items: [
        {
          id: 501,
          orgid: 29,
          task_no: 'inspect-20260420-02',
          status: 'pending',
          total_scanned: 0,
          hit_articles: 0,
          hit_count: 0,
          creator_name: 'operator',
          created_at: '2026-04-20T12:00:00Z'
        },
        {
          id: 502,
          orgid: 29,
          task_no: 'inspect-20260420-01',
          status: 'running',
          total_scanned: 42,
          hit_articles: 4,
          hit_count: 8,
          creator_name: 'operator',
          created_at: '2026-04-20T12:00:00Z'
        }
      ]
    });
    mockedDeleteTask.mockResolvedValue({ id: 501 } as never);
  });

  it('shows task result navigation and only allows deleting pending tasks', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.getByText('inspect-20260420-02')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '任务列表' })).toBeInTheDocument();
    expect(screen.queryByText('按任务编号与执行状态浏览当前批次。')).not.toBeInTheDocument();
    expect(screen.queryByText('更快筛出进行中的批次或定位单个任务编号。')).not.toBeInTheDocument();
    await user.click(screen.getByRole('link', { name: '新建任务' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/tasks/new');
    });

    await user.click(screen.getAllByRole('link', { name: '运行结果' })[0]);
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/tasks/501/results');
    });

    expect(screen.getByText('已执行不可删')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '删除任务' }));
    await user.click(await screen.findByRole('button', { name: '确认删除' }));
    expect(mockedDeleteTask).toHaveBeenCalledWith(501);
    expect(screen.queryByText('统一发起巡检任务，查看执行状态、扫描规模与命中情况。')).not.toBeInTheDocument();
  });

  it('hydrates task filters from the URL and restores the same state after remount', async () => {
    const user = userEvent.setup();
    const firstRender = renderPage(['/tasks?task_no=inspect-20260420-01&status=running&page=2']);

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.getByLabelText('任务编号')).toHaveValue('inspect-20260420-01');
    await waitFor(() => {
      const lastCall = mockedListTasks.mock.calls.at(-1)?.[0];
      expect(lastCall).toMatchObject({
        page: 2,
        pageSize: 20,
        task_no: 'inspect-20260420-01',
        status: 'running'
      });
      expect(lastCall).not.toHaveProperty('orgid');
    });

    await user.clear(screen.getByLabelText('任务编号'));
    await user.type(screen.getByLabelText('任务编号'), 'inspect-20260420-02');
    await user.click(screen.getByRole('button', { name: /查询任务/ }));

    await waitFor(() => {
      const locationState = readLocationState();
      expect(locationState.pathname).toBe('/tasks');
      expect(locationState.searchParams.get('task_no')).toBe('inspect-20260420-02');
      expect(locationState.searchParams.get('status')).toBe('running');
    });

    firstRender.unmount();
    renderPage(['/tasks?task_no=inspect-20260420-02&status=running']);

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.getByLabelText('任务编号')).toHaveValue('inspect-20260420-02');
    await waitFor(() => {
      const lastCall = mockedListTasks.mock.calls.at(-1)?.[0];
      expect(lastCall).toMatchObject({
        page: 1,
        pageSize: 20,
        task_no: 'inspect-20260420-02',
        status: 'running'
      });
      expect(lastCall).not.toHaveProperty('orgid');
    });
  });
});
