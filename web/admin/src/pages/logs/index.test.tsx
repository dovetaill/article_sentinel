import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import LogsPage from './index';
import { listOperationLogs } from '../../services/logs';

const { mockedListOrgs } = vi.hoisted(() => ({
  mockedListOrgs: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: mockedListOrgs
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn()
}));

const mockedListOperationLogs = vi.mocked(listOperationLogs);

describe('LogsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cateid: 0, enabled: true, sort: 1 }
    ]);
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
          request_snapshot: '{"reason":"spam"}',
          created_at: '2026-04-20 17:00:00'
        }
      ]
    } as never);
  });

  it('filters logs and links task records into the task-results workspace', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <LogsPage />
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/offline by operator/i)).toBeInTheDocument();
    expect(screen.queryByText('查询任务执行、结果处置与请求快照。')).not.toBeInTheDocument();
    expect(screen.queryByText('按稿件、任务和操作人回看关键记录。')).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: '#501' })).toHaveAttribute('href', '/articles/501');
    expect(screen.getByRole('link', { name: '#77' })).toHaveAttribute('href', '/tasks/77/results');
    expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();

    await user.clear(screen.getByLabelText('文章编号'));
    await user.type(screen.getByLabelText('文章编号'), '501');
    await user.clear(screen.getByLabelText('任务编号'));
    await user.type(screen.getByLabelText('任务编号'), '77');
    await user.clear(screen.getByLabelText('操作人'));
    await user.type(screen.getByLabelText('操作人'), 'alice');
    await user.click(screen.getByRole('button', { name: '查询日志' }));

    await waitFor(() => {
      expect(mockedListOperationLogs).toHaveBeenLastCalledWith(
        expect.objectContaining({
          article_id: 501,
          task_id: 77,
          operator_name: 'alice'
        }),
      );
    });
  });
});
