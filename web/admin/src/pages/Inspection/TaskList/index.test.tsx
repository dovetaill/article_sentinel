import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import TaskListPage from './index';

describe('TaskListPage', () => {
  it('shows the task list filters and create button', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks']}>
          <TaskListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByLabelText('任务编号')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新建任务' })).toBeInTheDocument();
  });
});
