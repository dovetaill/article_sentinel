import { ConfigProvider } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import RectifyPage from './rectify';
import { getArticleDetail, rectifyArticle, republishArticle } from '../../services/articles';

const { mockedListOrgs } = vi.hoisted(() => ({
  mockedListOrgs: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: mockedListOrgs
}));

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  batchIgnoreResults: vi.fn(),
  batchProcessResults: vi.fn()
}));

vi.mock('../../services/articles', () => ({
  listArticles: vi.fn(),
  getArticleDetail: vi.fn(),
  offlineArticle: vi.fn(),
  republishArticle: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedGetArticleDetail = vi.mocked(getArticleDetail);
const mockedRectifyArticle = vi.mocked(rectifyArticle);
const mockedRepublishArticle = vi.mocked(republishArticle);

describe('RectifyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cateid: 0, enabled: true, sort: 1 }
    ]);
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
    } as never);
    mockedRectifyArticle.mockResolvedValue([] as never);
    mockedRepublishArticle.mockResolvedValue({ article_id: 501, status: 'success' } as never);
  });

  it('loads from article detail and preserves untouched fields on save', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501/rectify']}>
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '内容整改' })).toBeInTheDocument();
    expect(screen.queryByText(/围绕当前稿件进行标题、摘要与正文修订。/)).not.toBeInTheDocument();
    expect(screen.queryByText('在保留事实准确性的前提下完成风险修订。')).not.toBeInTheDocument();
    expect(screen.queryByText('请在保持稿件主旨准确的前提下，对存在风险的标题、摘要与正文进行审慎修订，避免再次触发同类规则。')).not.toBeInTheDocument();
    expect(screen.queryByText('对照原始标题、摘要与正文，确保修改范围可控。')).not.toBeInTheDocument();
    expect(screen.queryByText('办理提示')).not.toBeInTheDocument();

    await user.clear(screen.getByLabelText('整改标题'));
    await user.type(screen.getByLabelText('整改标题'), 'New title');
    await user.clear(screen.getByLabelText('整改摘要'));
    await user.type(screen.getByLabelText('整改摘要'), 'New summary');
    await user.click(screen.getByRole('button', { name: '保存整改' }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenCalledWith(501, {
        orgid: 29,
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

  it('republishes after saving when submitting for review', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501/rectify']}>
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '保存并提交复核' }));

    await waitFor(() => {
      expect(mockedRepublishArticle).toHaveBeenCalledWith(501, {
        orgid: 29,
        task_id: 77,
        result_id: 9001
      });
    });
  });

  it('submits edited html source from the embedded editor', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501/rectify']}>
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: 'HTML源码' }));
    const sourceEditor = screen.getByRole('textbox', { name: '整改正文 HTML源码' });
    await user.clear(sourceEditor);
    await user.type(sourceEditor, '<section><p>Source body</p></section>');
    await user.click(screen.getByRole('button', { name: '保存整改' }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenLastCalledWith(501, expect.objectContaining({
        body: '<section><p>Source body</p></section>'
      }));
    });
  });

  it('keeps visual editing and html source in sync', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501/rectify']}>
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();

    const visualEditor = screen.getByRole('textbox', { name: '整改正文 可视化编辑' });
    visualEditor.innerHTML = '<p>Visual body</p><p>Second paragraph</p>';
    fireEvent.input(visualEditor);

    await user.click(screen.getByRole('tab', { name: 'HTML源码' }));
    expect(screen.getByRole('textbox', { name: '整改正文 HTML源码' })).toHaveValue('<p>Visual body</p><p>Second paragraph</p>');

    await user.click(screen.getByRole('button', { name: '保存整改' }));
    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenLastCalledWith(501, expect.objectContaining({
        body: '<p>Visual body</p><p>Second paragraph</p>'
      }));
    });
  });
});
