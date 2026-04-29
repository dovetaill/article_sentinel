import { ConfigProvider } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useEffect } from 'react';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { WorkbenchProvider } from '../../workbench/provider';
import { getWorkbenchSessionKey } from '../../workbench/session';
import { useWorkbench } from '../../workbench/use-workbench';
import RectifyPage from './rectify';
import { getArticleDetail, rectifyArticle, republishArticle } from '../../services/articles';

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
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

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function SeedWorkbenchTabs({ hrefs }: { hrefs: string[] }) {
  const { openTab } = useWorkbench();

  useEffect(() => {
    hrefs.forEach((href) => openTab(href));
  }, [hrefs, openTab]);

  return null;
}

function WorkbenchCycleControls() {
  const { activateTab } = useWorkbench();

  return (
    <>
      <button type="button" onClick={() => activateTab('/tasks')}>
        切换到任务列表
      </button>
      <button type="button" onClick={() => activateTab('article:501:rectify')}>
        切回整改页
      </button>
    </>
  );
}

function renderWorkbenchPage(initialEntries: string[] = ['/articles/501/rectify']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <WorkbenchCycleControls />
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
              <Route path="/tasks" element={<div>任务列表探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('RectifyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
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

  it('returns to the existing source tab before falling back to the return_to query target', async () => {
    window.sessionStorage.setItem(
      getWorkbenchSessionKey(29),
      JSON.stringify({
        orgId: 29,
        activeKey: '/articles',
        tabs: [
          {
            key: '/tasks',
            pathname: '/tasks',
            search: '',
            title: '检测任务',
            closable: false,
            keepAlive: false,
            orgId: 29
          },
          {
            key: '/articles',
            pathname: '/articles',
            search: '?page=2',
            title: '文稿中心',
            closable: true,
            keepAlive: false,
            orgId: 29
          }
        ]
      }),
    );

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/articles/501/rectify?return_to=%2Farticles%3Fview%3Dsummary']}>
          <OrgProvider>
            <WorkbenchProvider>
              <Routes>
                <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
                <Route path="/articles" element={<div>文稿列表探针</div>} />
              </Routes>
              <LocationProbe />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('link', { name: '返回上一页' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles?page=2');
    });
  });

  it('preserves unsaved rectify drafts across a workbench deactivate/reactivate cycle', async () => {
    const user = userEvent.setup();

    renderWorkbenchPage();

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();

    await user.clear(screen.getByLabelText('整改标题'));
    await user.type(screen.getByLabelText('整改标题'), 'Draft title');
    await user.clear(screen.getByLabelText('整改摘要'));
    await user.type(screen.getByLabelText('整改摘要'), 'Draft summary');

    await user.click(screen.getByRole('button', { name: '切换到任务列表' }));
    expect(await screen.findByText('任务列表探针')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '切回整改页' }));
    expect(await screen.findByText(/old title/i)).toBeInTheDocument();
    expect(screen.getByLabelText('整改标题')).toHaveValue('Draft title');
    expect(screen.getByLabelText('整改摘要')).toHaveValue('Draft summary');
  });
});
