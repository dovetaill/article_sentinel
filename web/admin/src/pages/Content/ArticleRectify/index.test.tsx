import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  mockedGetArticleDetail,
  mockedRectifyArticle,
  mockedRepublishArticle
} = vi.hoisted(() => ({
  mockedGetArticleDetail: vi.fn(),
  mockedRectifyArticle: vi.fn(),
  mockedRepublishArticle: vi.fn()
}));

vi.mock('@umijs/max', () => ({
  useModel: () => ({
    initialState: {
      currentUser: { orgid: 29, orgname: '一县一端' },
      currentOrgId: 29,
      currentOrgName: '一县一端'
    }
  })
}));

vi.mock('@/services/articles', () => ({
  listArticles: vi.fn(),
  getArticleDetail: mockedGetArticleDetail,
  offlineArticle: vi.fn(),
  republishArticle: mockedRepublishArticle,
  rectifyArticle: mockedRectifyArticle
}));

import ArticleRectifyPage from './index';

function renderPage(initialEntries: string[] = ['/content/articles/501/rectify?task_id=77&result_id=9001']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path="/content/articles/:articleId/rectify" element={<ArticleRectifyPage />} />
        </Routes>
      </MemoryRouter>
    </ConfigProvider>
  );
}

describe('ArticleRectifyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
    window.sessionStorage.clear();

    mockedGetArticleDetail.mockResolvedValue({
      id: 501,
      orgid: 29,
      title: 'Old title',
      short_title: 'Old short',
      rich_title: '<strong>Old rich</strong>',
      keyword: 'old-keyword',
      desc: 'Old summary',
      body: '<p>Old body</p>',
      state: 8,
      latest_task_id: 77,
      latest_result_id: 9001
    });
    mockedRectifyArticle.mockResolvedValue([]);
    mockedRepublishArticle.mockResolvedValue({
      article_id: 501,
      status: 'success'
    });
  });

  it('renders the rectify form and original article panel', async () => {
    const { container } = renderPage();

    expect(await screen.findByLabelText('整改标题')).toBeInTheDocument();
    expect(screen.getByLabelText('整改摘要')).toBeInTheDocument();
    expect(screen.getByText('原稿对照')).toBeInTheDocument();
    expect(screen.getByText('Old title')).toBeInTheDocument();
    expect(container.querySelectorAll('.admin-summary-card.admin-surface-panel')).toHaveLength(4);
    expect(container.querySelector('.rectify-layout__main.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.rectify-layout__side.admin-surface-panel')).toBeInTheDocument();
  });

  it('keeps a local draft while editing', async () => {
    const user = userEvent.setup();

    const firstRender = renderPage();

    const titleInput = await screen.findByLabelText('整改标题');
    await user.clear(titleInput);
    await user.type(titleInput, '新标题');

    firstRender.unmount();

    renderPage();

    expect(await screen.findByLabelText('整改标题')).toHaveValue('新标题');
  });

  it('saves rectification while preserving untouched article fields', async () => {
    const user = userEvent.setup();

    renderPage();

    await user.clear(await screen.findByLabelText('整改标题'));
    await user.type(screen.getByLabelText('整改标题'), 'New title');
    await user.clear(screen.getByLabelText('整改摘要'));
    await user.type(screen.getByLabelText('整改摘要'), 'New summary');
    await user.click(screen.getByRole('button', { name: '保存整改' }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenCalledWith(501, {
        task_id: 77,
        result_id: 9001,
        title: 'New title',
        short_title: 'Old short',
        rich_title: '<strong>Old rich</strong>',
        keyword: 'old-keyword',
        desc: 'New summary',
        body: '<p>Old body</p>'
      });
    });
  });

  it('saves and submits for review', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByLabelText('整改标题')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '保存并提交复核' }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenCalled();
      expect(mockedRepublishArticle).toHaveBeenCalledWith(501, {
        task_id: 77,
        result_id: 9001
      });
    });
  });
});
