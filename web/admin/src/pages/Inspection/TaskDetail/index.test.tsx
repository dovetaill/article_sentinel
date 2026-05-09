import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import TaskDetailPage from './index';

const { mockedListResults } = vi.hoisted(() => ({
  mockedListResults: vi.fn()
}));

vi.mock('@/services/tasks', () => ({
  getTaskDetail: vi.fn().mockResolvedValue({
    id: 77,
    orgid: 29,
    task_no: 'inspect-20260420-01',
    status: 'running',
    total_scanned: 42,
    hit_articles: 4,
    hit_count: 8,
    creator_name: 'operator',
    created_at: '2026-04-20 12:00:00',
    rule_snapshot: '{"keywords":["spam","scam"]}',
    request_snapshot: '{"title_like":"Alpha"}'
  }),
  listTasks: vi.fn(),
  createTask: vi.fn(),
  deleteTask: vi.fn()
}));

vi.mock('@/services/results', () => ({
  listResults: mockedListResults,
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('@/services/logs', () => ({
  listOperationLogs: vi.fn().mockResolvedValue({
    page: 1,
    pageSize: 20,
    total: 0,
    items: []
  }),
  listArticleOperationLogs: vi.fn(),
  listArticleFieldChanges: vi.fn()
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('TaskDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 29,
          task_id: 77,
          article_id: 501,
          article_title: '县域融媒今日要闻',
          risk_level: 'high',
          disposition_status: 'pending',
          hit_count: 1,
          snippet: '命中片段'
        }
      ]
    });
  });

  it('renders task detail tabs for hit results and snapshots', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks/77']}>
          <Routes>
            <Route path="/inspection/tasks/:taskId" element={<TaskDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('tab', { name: '命中结果' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '规则快照' })).toBeInTheDocument();
  });

  it('keeps the task detail as return_to when drilling into an article', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/tasks/77']}>
          <Routes>
            <Route path="/inspection/tasks/:taskId" element={<TaskDetailPage />} />
          </Routes>
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '县域融媒今日要闻' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent(
        '/content/articles/501?return_to=%2Finspection%2Ftasks%2F77'
      );
    });
  });
});
