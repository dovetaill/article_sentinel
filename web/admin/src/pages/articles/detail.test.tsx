import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ArticleDetailPage from './detail';
import { listArticleFieldChanges, listArticleOperationLogs } from '../../services/logs';
import { getResultDetail, listResults } from '../../services/results';

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

const mockedListResults = vi.mocked(listResults);
const mockedGetResultDetail = vi.mocked(getResultDetail);
const mockedListArticleOperationLogs = vi.mocked(listArticleOperationLogs);
const mockedListArticleFieldChanges = vi.mocked(listArticleFieldChanges);

describe('ArticleDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
          article_state: 9,
          risk_level: 'high',
          suggest_action: 'offline',
          disposition_status: 'pending',
          hit_count: 3
        }
      ]
    } as never);

    mockedGetResultDetail.mockResolvedValue({
      id: 11,
      orgid: 100,
      task_id: 77,
      article_id: 501,
      article_title: 'Spam alert',
      article_state: 9,
      risk_level: 'high',
      suggest_action: 'offline',
      disposition_status: 'pending',
      hit_count: 3,
      article_body: '<p>spam body</p>',
      hits: [
        {
          id: 1,
          field_name: 'title',
          keyword_text: 'spam',
          snippet: 'spam alert keeps repeating',
          matched_text: 'spam',
          risk_level: 'high'
        }
      ],
      operation_logs: [],
      field_changes: []
    } as never);

    mockedListArticleOperationLogs.mockResolvedValue({
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
          before_state: '9',
          after_state: '8',
          summary: 'Article sent offline',
          operator_name: 'auditor',
          created_at: '2026-04-20 16:00:00'
        }
      ]
    } as never);

    mockedListArticleFieldChanges.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 3,
          orgid: 100,
          article_id: 501,
          field_name: 'title',
          before_value: 'old',
          after_value: 'new',
          diff_summary: 'title updated',
          operator_name: 'auditor',
          created_at: '2026-04-20 16:10:00'
        }
      ]
    } as never);
  });

  it('renders the full article detail workspace', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/articles/501']}>
          <Routes>
            <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();
    expect(screen.getAllByText('文稿编号').length).toBeGreaterThan(0);
    expect(screen.getByRole('tab', { name: '命中记录' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '进入整改' })).toHaveAttribute('href', '/articles/501/rectify');
    await user.click(screen.getByRole('tab', { name: '操作记录' }));
    expect(screen.getByText(/article sent offline/i)).toBeInTheDocument();
  });
});
