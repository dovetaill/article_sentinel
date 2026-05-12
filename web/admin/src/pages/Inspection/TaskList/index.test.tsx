import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, vi } from 'vitest';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import TaskListPage from './index';

const { mockedListTasks } = vi.hoisted(() => ({
  mockedListTasks: vi.fn()
}));

vi.mock('@/services/tasks', () => ({
  listTasks: mockedListTasks,
  deleteTask: vi.fn()
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

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

  it('restores task detail entry points from task number and row actions', async () => {
    const user = userEvent.setup();
    mockedListTasks.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 77,
          orgid: 29,
          task_no: 'inspect-20260420-01',
          status: 'success',
          total_scanned: 42,
          hit_count: 8,
          creator_name: 'operator',
          created_at: '2026-04-20 12:00:00'
        }
      ]
    });

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks']}>
          <TaskListPage />
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('button', { name: 'inspect-20260420-01' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '查看任务' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'inspect-20260420-01' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/inspection/tasks/77');
    });

    await user.click(screen.getByRole('button', { name: '查看任务' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/inspection/tasks/77');
    });
  });
});
