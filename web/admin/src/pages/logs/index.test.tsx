import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import LogsPage from './index';
import { listOperationLogs } from '../../services/logs';
import { WorkbenchProvider } from '../../workbench/provider';

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn()
}));

const mockedListOperationLogs = vi.mocked(listOperationLogs);

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/logs']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <LogsPage />
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

describe('LogsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOperationLogs.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 91,
          orgid: 29,
          article_id: 501,
          task_id: 77,
          operation_type: 'offline',
          before_state: 'online',
          after_state: 'offline',
          summary: 'offline by operator',
          operator_name: 'alice',
          request_snapshot: '{"include_body":true,"reason":"spam"}',
          created_at: '2026-04-20 17:00:00'
        }
      ]
    } as never);
  });

  function expectLastLogQuery(expected: Record<string, unknown>) {
    const lastCall = mockedListOperationLogs.mock.calls.at(-1)?.[0];
    expect(lastCall).toMatchObject(expected);
    expect(lastCall).not.toHaveProperty('orgid');
  }

  it('filters logs and links task records into the task-results workspace', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText(/offline by operator/i)).toBeInTheDocument();
    expect(screen.queryByText('查询任务执行、结果处置与请求快照。')).not.toBeInTheDocument();
    expect(screen.queryByText('按稿件、任务和操作人回看关键记录。')).not.toBeInTheDocument();
    expect(screen.queryByText('快速串联任务、文章与操作人三个维度的记录。')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: '#501' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501?return_to=%2Flogs');
    });

    await user.click(screen.getByRole('link', { name: '#77' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/tasks/77/results');
    });

    await user.clear(screen.getByLabelText('文章编号'));
    await user.type(screen.getByLabelText('文章编号'), '501');
    await user.clear(screen.getByLabelText('任务编号'));
    await user.type(screen.getByLabelText('任务编号'), '77');
    await user.clear(screen.getByLabelText('操作人'));
    await user.type(screen.getByLabelText('操作人'), 'alice');
    await user.click(screen.getByRole('button', { name: '查询日志' }));

    await waitFor(() => {
      expectLastLogQuery({
        article_id: 501,
        task_id: 77,
        operator_name: 'alice'
      });
    });

    await user.click(screen.getByRole('button', { name: '查看快照' }));
    expect(screen.getByRole('dialog', { name: '请求快照' })).toBeInTheDocument();
    expect(screen.getByText(/reason/i)).toBeInTheDocument();
    expect(screen.getByText(/spam/i)).toBeInTheDocument();
    expect(screen.queryByText(/include_body/i)).not.toBeInTheDocument();
  });

  it('hydrates log filters from the URL and restores the same state after remount', async () => {
    const user = userEvent.setup();
    const firstRender = renderPage(['/logs?article_id=501&task_id=77&operator_name=alice&page=2']);

    expect(await screen.findByText(/offline by operator/i)).toBeInTheDocument();
    expect(screen.getByLabelText('文章编号')).toHaveValue('501');
    expect(screen.getByLabelText('任务编号')).toHaveValue('77');
    expect(screen.getByLabelText('操作人')).toHaveValue('alice');
    await waitFor(() => {
      expectLastLogQuery({
        page: 2,
        pageSize: 20,
        article_id: 501,
        task_id: 77,
        operator_name: 'alice'
      });
    });

    await user.clear(screen.getByLabelText('文章编号'));
    await user.type(screen.getByLabelText('文章编号'), '808');
    await user.clear(screen.getByLabelText('任务编号'));
    await user.type(screen.getByLabelText('任务编号'), '99');
    await user.clear(screen.getByLabelText('操作人'));
    await user.type(screen.getByLabelText('操作人'), 'bob');
    await user.click(screen.getByRole('button', { name: '查询日志' }));

    await waitFor(() => {
      const locationState = readLocationState();
      expect(locationState.pathname).toBe('/logs');
      expect(locationState.searchParams.get('article_id')).toBe('808');
      expect(locationState.searchParams.get('task_id')).toBe('99');
      expect(locationState.searchParams.get('operator_name')).toBe('bob');
    });

    firstRender.unmount();
    renderPage(['/logs?article_id=808&task_id=99&operator_name=bob']);

    expect(await screen.findByText(/offline by operator/i)).toBeInTheDocument();
    expect(screen.getByLabelText('文章编号')).toHaveValue('808');
    expect(screen.getByLabelText('任务编号')).toHaveValue('99');
    expect(screen.getByLabelText('操作人')).toHaveValue('bob');
    await waitFor(() => {
      expectLastLogQuery({
        page: 1,
        pageSize: 20,
        article_id: 808,
        task_id: 99,
        operator_name: 'bob'
      });
    });
  });
});
