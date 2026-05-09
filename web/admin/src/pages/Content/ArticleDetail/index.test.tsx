import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ArticleDetailPage from './index';

const {
  mockedGetArticleDetail,
  mockedListArticleOperationLogs,
  mockedListArticleFieldChanges,
  mockedGetResultDetail
} = vi.hoisted(() => ({
  mockedGetArticleDetail: vi.fn(),
  mockedListArticleOperationLogs: vi.fn(),
  mockedListArticleFieldChanges: vi.fn(),
  mockedGetResultDetail: vi.fn()
}));

vi.mock('@/services/articles', () => ({
  getArticleDetail: mockedGetArticleDetail,
  listArticles: vi.fn(),
  offlineArticle: vi.fn(),
  republishArticle: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('@/services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: mockedGetResultDetail,
  batchOfflineResults: vi.fn(),
  batchIgnoreResults: vi.fn(),
  batchProcessResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('@/services/logs', () => ({
  listOperationLogs: vi.fn(),
  listArticleOperationLogs: mockedListArticleOperationLogs,
  listArticleFieldChanges: mockedListArticleFieldChanges
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('ArticleDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetArticleDetail.mockResolvedValue({
      id: 501,
      orgid: 29,
      title: '县域融媒今日要闻',
      short_title: '今日要闻',
      rich_title: '<p>县域融媒今日要闻 rich</p>',
      keyword: 'spam',
      desc: '真实摘要',
      body: '<p>真实正文</p>',
      state: 9,
      latest_risk_level: 'high',
      latest_task_id: 208,
      latest_result_id: 11,
      latest_disposition_status: 'pending',
      latest_suggest_action: 'offline'
    });
    mockedGetResultDetail.mockResolvedValue({
      id: 11,
      orgid: 29,
      task_id: 208,
      article_id: 501,
      article_title: '县域融媒今日要闻',
      risk_level: 'high',
      suggest_action: 'offline',
      disposition_status: 'pending',
      hit_count: 1,
      hits: [],
      operation_logs: [],
      field_changes: []
    });
    mockedListArticleOperationLogs.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 0,
      items: []
    });
    mockedListArticleFieldChanges.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 0,
      items: []
    });
  });

  it('renders article detail actions and tabs', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/content/articles/501']}>
          <Routes>
            <Route path="/content/articles/:articleId" element={<ArticleDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('button', { name: '进入整改' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '命中记录' })).toBeInTheDocument();
  });

  it('keeps the article detail visible when audit side requests fail', async () => {
    mockedListArticleOperationLogs.mockRejectedValue(new Error('日志接口失败'));
    mockedListArticleFieldChanges.mockRejectedValue(new Error('变更接口失败'));

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/content/articles/501']}>
          <Routes>
            <Route path="/content/articles/:articleId" element={<ArticleDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('真实正文')).toBeInTheDocument();
    expect(screen.getByText('文稿详情')).toBeInTheDocument();
    expect(screen.queryByText('未查询到该文章的中心数据。')).not.toBeInTheDocument();
  });

  it('navigates back with return_to semantics from the query string', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/content/articles/501?return_to=%2Finspection%2Fresults%3Fpage%3D2']}>
          <Routes>
            <Route path="/content/articles/:articleId" element={<ArticleDetailPage />} />
            <Route path="/inspection/results" element={<div>结果页探针</div>} />
          </Routes>
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('button', { name: '返回上一页' })).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '返回上一页' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/inspection/results?page=2');
    });
  });
});
