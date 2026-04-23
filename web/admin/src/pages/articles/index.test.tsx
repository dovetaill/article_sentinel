import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ArticlesPage from './index';
import { listArticles } from '../../services/articles';

vi.mock('../../services/articles', () => ({
  listArticles: vi.fn()
}));

const mockedListArticles = vi.mocked(listArticles);

describe('ArticlesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListArticles.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          article_id: 501,
          article_title: 'Spam alert',
          article_state: 9,
          risk_level: 'high',
          disposition_status: 'pending',
          hit_count: 3,
          latest_task_id: 208,
          latest_operator_name: '值班员乙',
          latest_action_at: '2026-04-20 12:00:00'
        }
      ]
    } as never);
  });

  it('renders article-centric inspection rows', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter>
          <ArticlesPage />
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(screen.queryByText('按稿件维度查看巡检命中、处置状态与最近任务，便于持续跟进单篇稿件。')).not.toBeInTheDocument();
    expect(await screen.findByText('Spam alert')).toBeInTheDocument();
    expect(screen.getByText('#501')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/articles/501');
  });
});
