import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import TaskCreatePage from './index';

vi.mock('@/services/keywords', () => ({
  listKeywords: vi.fn().mockResolvedValue({
    page: 1,
    pageSize: 20,
    total: 0,
    items: []
  })
}));

vi.mock('@/services/tasks', () => ({
  createTask: vi.fn(),
  deleteTask: vi.fn(),
  getTaskDetail: vi.fn(),
  listTasks: vi.fn()
}));

describe('TaskCreatePage', () => {
  it('requires at least one rule before submitting a task', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks/create']}>
          <TaskCreatePage />
        </MemoryRouter>
      </ConfigProvider>
    );

    await user.click(screen.getByRole('button', { name: '提交任务' }));

    expect(await screen.findByText('请先选择至少一条规则')).toBeInTheDocument();
  });
});
