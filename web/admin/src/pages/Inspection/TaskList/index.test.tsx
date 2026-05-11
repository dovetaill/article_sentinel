import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { beforeEach, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import TaskListPage from './index';

const { mockedListTasks } = vi.hoisted(() => ({
  mockedListTasks: vi.fn()
}));

vi.mock('@/services/tasks', () => ({
  listTasks: mockedListTasks,
  deleteTask: vi.fn()
}));

describe('TaskListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListTasks.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 0,
      items: []
    });
  });

  it('shows the task list filters and create button', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks']}>
          <TaskListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByLabelText('任务编号')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新建任务' })).toBeInTheDocument();
    expect(container.querySelectorAll('.admin-summary-card.admin-surface-panel')).toHaveLength(4);
    expect(container.querySelector('.admin-filter-card.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-surface-panel')).toBeInTheDocument();
  });

  it('keeps the task list page on light surfaces after the shell refactor', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks']}>
          <TaskListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    await screen.findByLabelText('任务编号');

    expect(container.querySelector('.admin-domain-page__head.admin-light-surface')).toBeInTheDocument();
    expect(container.querySelectorAll('.admin-summary-card.admin-light-surface')).toHaveLength(4);
    expect(container.querySelector('.admin-filter-card.admin-light-surface')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-light-surface')).toBeInTheDocument();
  });
});
