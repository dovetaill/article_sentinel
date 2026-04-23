import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TasksPage from './index';
import { listTasks } from '../../services/tasks';

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

const mockedListTasks = vi.mocked(listTasks);

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
  });

  it('renders task status tags and routes detail action to the full detail page', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter>
          <TasksPage />
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.queryByText('统一发起巡检任务，查看执行状态、扫描规模与命中情况。')).not.toBeInTheDocument();
    expect(screen.getAllByText('执行中').length).toBeGreaterThan(0);
    expect(screen.getByRole('link', { name: '新建任务' })).toHaveAttribute('href', '/tasks/new');
    expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/tasks/77');
    expect(screen.queryByText(/spam keyword/i)).not.toBeInTheDocument();
  });
});
