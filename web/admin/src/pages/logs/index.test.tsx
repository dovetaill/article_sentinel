import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import LogsPage from './index';
import { listOperationLogs } from '../../services/logs';

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn()
}));

const mockedListOperationLogs = vi.mocked(listOperationLogs);

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
          orgid: 100,
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

  it('filters by article, operator, and task', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <LogsPage />
      </ConfigProvider>,
    );

    expect(await screen.findByText(/offline by operator/i)).toBeInTheDocument();

    await user.clear(screen.getByLabelText(/article id/i));
    await user.type(screen.getByLabelText(/article id/i), '501');
    await user.clear(screen.getByLabelText(/task id/i));
    await user.type(screen.getByLabelText(/task id/i), '77');
    await user.clear(screen.getByLabelText(/operator/i));
    await user.type(screen.getByLabelText(/operator/i), 'alice');
    await user.click(screen.getByRole('button', { name: /search logs/i }));

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
