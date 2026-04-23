import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TaskDetailPage from './detail';
import { listOperationLogs } from '../../services/logs';
import { listResults } from '../../services/results';
import { getTaskDetail } from '../../services/tasks';

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn(),
  listArticleOperationLogs: vi.fn(),
  listArticleFieldChanges: vi.fn()
}));

const mockedGetTaskDetail = vi.mocked(getTaskDetail);
const mockedListResults = vi.mocked(listResults);
const mockedListOperationLogs = vi.mocked(listOperationLogs);

describe('TaskDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetTaskDetail.mockResolvedValue({
      id: 77,
      orgid: 100,
      task_no: 'inspect-20260420-01',
      status: 'running',
      total_scanned: 42,
      hit_articles: 4,
      hit_count: 8,
      creator_name: 'operator',
      created_at: '2026-04-20 12:00:00',
      rule_snapshot: '{"keywords":["spam","scam"]}',
      request_snapshot: '{"include_body":true,"title_like":"Alpha"}'
    } as never);

    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 100,
          task_id: 77,
          article_id: 501,
          article_title: 'Spam alert',
          risk_level: 'high',
          suggest_action: 'offline',
          disposition_status: 'pending',
          hit_count: 3,
          matched_keyword: 'spam',
          snippet: 'This spam alert keeps repeating'
        }
      ]
    } as never);

    mockedListOperationLogs.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 9,
          orgid: 100,
          article_id: 501,
          task_id: 77,
          operation_type: 'offline',
          before_state: 'online',
          after_state: 'offline',
          summary: 'Task reviewed by auditor',
          operator_name: 'auditor',
          created_at: '2026-04-20 16:00:00'
        }
      ]
    } as never);
  });

  it('renders the full task detail workspace with tabs and linked hit results', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks/77']}>
          <Routes>
            <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect((await screen.findAllByText('inspect-20260420-01')).length).toBeGreaterThan(0);
    expect(screen.getByRole('link', { name: '返回任务列表' })).toHaveAttribute('href', '/tasks');
    expect(screen.getByRole('tab', { name: '命中结果' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '规则快照' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '请求快照' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '关联日志' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Spam alert' })).toHaveAttribute('href', '/articles/501');

    await user.click(screen.getByRole('tab', { name: '关联日志' }));
    expect(screen.getAllByText(/task reviewed by auditor/i).length).toBeGreaterThan(0);
  });
});
