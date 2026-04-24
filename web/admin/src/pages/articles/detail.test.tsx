import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import ArticleDetailPage from './detail';

const {
  mockedListOrgs,
  mockedGetArticleDetail,
  mockedOfflineArticle,
  mockedRepublishArticle,
  mockedListResults,
  mockedGetResultDetail,
  mockedListArticleOperationLogs,
  mockedListArticleFieldChanges
} = vi.hoisted(() => ({
  mockedListOrgs: vi.fn(),
  mockedGetArticleDetail: vi.fn(),
  mockedOfflineArticle: vi.fn(),
  mockedRepublishArticle: vi.fn(),
  mockedListResults: vi.fn(),
  mockedGetResultDetail: vi.fn(),
  mockedListArticleOperationLogs: vi.fn(),
  mockedListArticleFieldChanges: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: mockedListOrgs
}));

vi.mock('../../services/articles', () => ({
  getArticleDetail: mockedGetArticleDetail,
  offlineArticle: mockedOfflineArticle,
  republishArticle: mockedRepublishArticle
}));

vi.mock('../../services/results', () => ({
  listResults: mockedListResults,
  getResultDetail: mockedGetResultDetail,
  batchOfflineResults: vi.fn(),
  batchIgnoreResults: vi.fn(),
  batchProcessResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn(),
  listArticleOperationLogs: mockedListArticleOperationLogs,
  listArticleFieldChanges: mockedListArticleFieldChanges
}));

describe('ArticleDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cateid: 0, enabled: true, sort: 1 }
    ]);
    mockedGetArticleDetail.mockResolvedValue({
      id: 501,
      orgid: 29,
      title: '县域融媒今日要闻',
      short_title: '今日要闻',
      rich_title: '县域融媒今日要闻 rich',
      keyword: 'spam',
      desc: '真实摘要',
      body: '<p>真实正文</p>',
      state: 9,
      latest_risk_level: 'high',
      latest_task_id: 208,
      latest_disposition_status: 'pending',
      latest_suggest_action: 'offline'
    });
    mockedListResults.mockResolvedValue({ page: 1, pageSize: 20, total: 0, items: [] });
    mockedGetResultDetail.mockResolvedValue(null);
    mockedListArticleOperationLogs.mockResolvedValue({ page: 1, pageSize: 20, total: 1, items: [] });
    mockedListArticleFieldChanges.mockResolvedValue({ page: 1, pageSize: 20, total: 0, items: [] });
  });

  it('renders the real article body, latest inspect summary, and lifecycle actions', async () => {
    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501']}>
            <Routes>
              <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByText('真实摘要')).toBeInTheDocument();
    expect(screen.getByText('真实正文')).toBeInTheDocument();
    expect(screen.getByText(/最近巡检摘要/)).toBeInTheDocument();
    expect(screen.getByText(/最新风险/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '进入整改' })).toHaveAttribute('href', '/articles/501/rectify');
    expect(screen.getByRole('button', { name: '下线处置' })).toBeInTheDocument();
    expect(screen.queryByText('集中查看单篇文稿的命中情况、正文快照、处置记录与整改入口。')).not.toBeInTheDocument();
  });
});
