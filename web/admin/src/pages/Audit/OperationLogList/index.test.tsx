import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { mockedListOperationLogs } = vi.hoisted(() => ({
  mockedListOperationLogs: vi.fn()
}));

vi.mock('@/services/logs', () => ({
  listOperationLogs: mockedListOperationLogs,
  listArticleOperationLogs: vi.fn(),
  listArticleFieldChanges: vi.fn()
}));

import OperationLogListPage from './index';

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('OperationLogListPage', () => {
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
    });
  });

  it('renders log filters and the snapshot action', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/audit/logs']}>
          <OperationLogListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByLabelText('文章编号')).toBeInTheDocument();
    expect(screen.getByLabelText('任务编号')).toBeInTheDocument();
    expect(screen.getByLabelText('操作人')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: '查看快照' })).toBeInTheDocument();
    expect(container.querySelectorAll('.admin-summary-card.admin-surface-panel')).toHaveLength(4);
    expect(container.querySelector('.admin-filter-card.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-surface-panel')).toBeInTheDocument();
  });

  it('hydrates URL filters and opens the request snapshot modal', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/audit/logs?article_id=501&task_id=77&operator_name=alice&page=2']}>
          <OperationLogListPage />
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('offline by operator')).toBeInTheDocument();
    expect(screen.getByLabelText('文章编号')).toHaveValue('501');
    expect(screen.getByLabelText('任务编号')).toHaveValue('77');
    expect(screen.getByLabelText('操作人')).toHaveValue('alice');

    await waitFor(() => {
      expect(mockedListOperationLogs).toHaveBeenLastCalledWith(
        expect.objectContaining({
          page: 2,
          pageSize: 20,
          article_id: 501,
          task_id: 77,
          operator_name: 'alice'
        })
      );
    });

    await user.click(screen.getByRole('button', { name: '查看快照' }));
    const dialog = screen.getByRole('dialog', { name: '请求快照' });
    expect(dialog).toBeInTheDocument();
    expect(document.querySelector('.admin-light-modal.admin-operation-log-modal')).toBeInTheDocument();
    expect(document.querySelector('.snapshot-viewer.admin-surface-inline')).toBeInTheDocument();
    expect(within(dialog).getByText(/reason/i)).toBeInTheDocument();
    expect(within(dialog).getByText(/spam/i)).toBeInTheDocument();
  });

  it('updates the audit URL filters and routes record links into the new workspaces', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/audit/logs']}>
          <OperationLogListPage />
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('offline by operator')).toBeInTheDocument();

    await user.clear(screen.getByLabelText('文章编号'));
    await user.type(screen.getByLabelText('文章编号'), '808');
    await user.clear(screen.getByLabelText('任务编号'));
    await user.type(screen.getByLabelText('任务编号'), '99');
    await user.clear(screen.getByLabelText('操作人'));
    await user.type(screen.getByLabelText('操作人'), 'bob');
    await user.click(screen.getByRole('button', { name: '查询日志' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent(
        '/audit/logs?article_id=808&task_id=99&operator_name=bob'
      );
    });

    await user.click(screen.getByRole('button', { name: '#501' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent(
        '/content/articles/501?return_to=%2Faudit%2Flogs%3Farticle_id%3D808%26task_id%3D99%26operator_name%3Dbob'
      );
    });
  });
});
