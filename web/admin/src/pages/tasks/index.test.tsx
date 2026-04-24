import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listOrgs } from '../../services/orgs';
import TasksPage from './index';
import { listTasks } from '../../services/tasks';

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
}));

const mockedListTasks = vi.mocked(listTasks);
const mockedListOrgs = vi.mocked(listOrgs);

describe('TasksPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      {
        id: 29,
        name: '一县一端',
        cateid: 0,
        enabled: true,
        sort: 1
      }
    ]);
    mockedListTasks.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 501,
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
  });

  it('shows the task-results entry as the primary row action', async () => {
    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter>
            <TasksPage />
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '任务列表' })).toBeInTheDocument();
    expect(screen.queryByText('按任务编号与执行状态浏览当前批次。')).not.toBeInTheDocument();
    expect(screen.queryByText('更快筛出进行中的批次或定位单个任务编号。')).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: '新建任务' })).toHaveAttribute('href', '/tasks/new');
    expect(screen.getByRole('link', { name: '运行结果' })).toHaveAttribute('href', '/tasks/501/results');
    expect(screen.queryByText('统一发起巡检任务，查看执行状态、扫描规模与命中情况。')).not.toBeInTheDocument();
  });
});
