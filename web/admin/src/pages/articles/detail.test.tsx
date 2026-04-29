import { ConfigProvider } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { WorkbenchProvider } from '../../workbench/provider';
import { useWorkbench } from '../../workbench/use-workbench';
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

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
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

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/articles/501']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <Routes>
              <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
              <Route path="/articles/:articleId/rectify" element={<div>整改页探针</div>} />
              <Route path="/results" element={<div>结果页探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function WorkbenchCycleControls() {
  const { activateTab } = useWorkbench();

  return (
    <>
      <button type="button" onClick={() => activateTab('/tasks')}>
        切换到任务列表
      </button>
      <button type="button" onClick={() => activateTab('article:501')}>
        切回文稿详情
      </button>
    </>
  );
}

function renderWorkbenchPage(initialEntries: string[] = ['/articles/501']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <WorkbenchCycleControls />
            <Routes>
              <Route path="/articles/:articleId" element={<ArticleDetailPage />} />
              <Route path="/articles/:articleId/rectify" element={<div>整改页探针</div>} />
              <Route path="/results" element={<div>结果页探针</div>} />
              <Route path="/tasks" element={<div>任务列表探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('ArticleDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.sessionStorage.clear();
    mockedGetArticleDetail.mockResolvedValue({
      id: 501,
      orgid: 29,
      title: '县域融媒今日要闻',
      short_title: '今日要闻',
      rich_title: '<p>县域融媒今日要闻 rich</p><img alt="富标题配图" src="https://example.com/rich.png" />',
      keyword: 'spam',
      desc: '真实摘要',
      body: '<p>真实正文</p><img alt="正文配图" src="https://example.com/body.png" />',
      thumbnail: 'https://example.com/thumb.png',
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

  it('renders the real article body, rich title preview, and opens rectify through the workbench route', async () => {
    renderPage();

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByText('真实摘要')).toBeInTheDocument();
    expect(screen.getByText('真实正文')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: '文稿封面' })).toHaveAttribute('src', 'https://example.com/thumb.png');
    expect(screen.getByRole('img', { name: '富标题配图' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: '正文配图' })).toBeInTheDocument();
    expect(screen.getByText(/最近巡检摘要/)).toBeInTheDocument();
    expect(screen.getByText(/最新风险/)).toBeInTheDocument();
    expect(screen.getByText('县域融媒今日要闻 rich')).toBeInTheDocument();
    expect(screen.queryByText('查看真实文稿原文，并附带最近一次巡检留痕作为参考。')).not.toBeInTheDocument();
    expect(screen.queryByText('结合最近一次巡检补充信息做出处置判断。')).not.toBeInTheDocument();
    expect(screen.queryByText('最近一次巡检任务')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '下线处置' })).not.toBeInTheDocument();
    expect(screen.queryByText('集中查看单篇文稿的命中情况、正文快照、处置记录与整改入口。')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('link', { name: '进入整改' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501/rectify?return_to=%2Farticles&task_id=208');
    });
  });

  it('preserves query-string return targets for workbench-aware back navigation', async () => {
    renderPage(['/articles/501?return_to=%2Fresults%3Fpage%3D2']);

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('link', { name: '返回上一页' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/results?page=2');
    });
  });

  it('restores the last local tab after a workbench deactivate/reactivate cycle', async () => {
    const user = userEvent.setup();

    renderWorkbenchPage();

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: '操作记录' }));
    expect(screen.getByText('暂无操作记录。')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '切换到任务列表' }));
    expect(await screen.findByText('任务列表探针')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '切回文稿详情' }));
    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByText('暂无操作记录。')).toBeInTheDocument();
  });
});
