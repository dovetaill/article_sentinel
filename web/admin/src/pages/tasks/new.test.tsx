import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import NewTaskPage from './new';
import { listKeywords } from '../../services/keywords';
import { createTask } from '../../services/tasks';

vi.mock('../../services/keywords', () => ({
  listKeywords: vi.fn()
}));

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

const mockedListKeywords = vi.mocked(listKeywords);
const mockedCreateTask = vi.mocked(createTask);

describe('NewTaskPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListKeywords.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 2,
      items: [
        { id: 7, name: 'spam', orgid: 100 },
        { id: 8, name: 'scam', orgid: 100 }
      ]
    } as never);
    mockedCreateTask.mockResolvedValue({ id: 88, task_no: 'inspect-20260420-88' } as never);
  });

  it('submits valid task form data', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <NewTaskPage />
      </ConfigProvider>,
    );

    await user.type(screen.getByLabelText(/orgid/i), '100');
    await user.selectOptions(screen.getByLabelText(/keyword set/i), '7');
    await user.click(screen.getByRole('switch', { name: /include body/i }));
    await user.type(screen.getByLabelText(/article id/i), '123');
    await user.type(screen.getByLabelText(/title like/i), 'Alpha');
    await user.click(screen.getByRole('button', { name: /launch inspection/i }));

    await waitFor(() => {
      expect(mockedCreateTask).toHaveBeenCalledWith(
        expect.objectContaining({
          orgid: 100,
          keyword_ids: [7],
          include_body: true,
          article_id: 123,
          title_like: 'Alpha'
        }),
      );
    });
  });
});
