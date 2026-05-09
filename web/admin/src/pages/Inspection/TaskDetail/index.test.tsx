import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import TaskDetailPage from './index';

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
  listResults: vi.fn().mockResolvedValue({
    page: 1,
    pageSize: 20,
    total: 0,
    items: []
  }),
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

describe('TaskDetailPage', () => {
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
});
