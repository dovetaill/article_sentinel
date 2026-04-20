import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TasksPage from './index';
import { getTaskDetail, listTasks } from '../../services/tasks';

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

const mockedListTasks = vi.mocked(listTasks);
const mockedGetTaskDetail = vi.mocked(getTaskDetail);

describe('TasksPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListTasks.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 77,
          orgid: 100,
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
    mockedGetTaskDetail.mockResolvedValue({
      id: 77,
      task_no: 'inspect-20260420-01',
      status: 'running',
      creator_name: 'operator',
      rule_snapshot: 'spam keyword'
    } as never);
  });

  it('renders task status tags and detail action', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <TasksPage />
      </ConfigProvider>,
    );

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.getByText('running')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /detail/i }));
    expect(await screen.findByText(/spam keyword/i)).toBeInTheDocument();
  });
});
