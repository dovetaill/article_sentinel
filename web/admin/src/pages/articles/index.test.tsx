import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { WorkbenchProvider } from '../../workbench/provider';
import ArticlesPage from './index';

const { mockedListArticles, mockedListOrgs } = vi.hoisted(() => ({
  mockedListArticles: vi.fn(),
  mockedListOrgs: vi.fn()
}));

vi.mock('../../services/articles', () => ({
  listArticles: mockedListArticles
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: mockedListOrgs
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/articles']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <ArticlesPage />
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function readLocationState() {
  const text = screen.getByTestId('location-probe').textContent ?? '';
  const [pathname, search = ''] = text.split('?');

  return {
    pathname,
    searchParams: new URLSearchParams(search)
  };
}

describe('ArticlesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 }
    ]);
    mockedListArticles.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 25,
      items: [
        {
          id: 501,
          orgid: 29,
          title: '县域融媒今日要闻',
          state: 9,
          latest_risk_level: 'high',
          latest_task_id: 208,
          latest_disposition_status: 'pending'
        },
        {
          id: 502,
          orgid: 29,
          title: '已下线稿件',
          state: 8
        }
      ]
    });
  });

  it('renders article search, pagination, and latest inspect enrichment', async () => {
    const user = userEvent.setup();

    mockedListArticles.mockImplementation(async (params) => {
      if (params.page === 2) {
        return {
          page: 2,
          pageSize: 20,
          total: 25,
          items: [
            {
              id: 777,
              orgid: 29,
              title: '第二页稿件',
              state: 9
            }
          ]
        };
      }

      if (params.article_id) {
        return {
          page: 1,
          pageSize: 20,
          total: 1,
          items: [
            {
              id: 901,
              orgid: 29,
              title: 'ID 精确命中文稿',
              state: 9
            }
          ]
        };
      }

      if (params.title) {
        return {
          page: 1,
          pageSize: 20,
          total: 1,
          items: [
            {
              id: 601,
              orgid: 29,
              title: '搜索命中文稿',
              state: 9
            }
          ]
        };
      }

      return {
        page: 1,
        pageSize: 20,
        total: 25,
        items: [
          {
            id: 501,
            orgid: 29,
            title: '县域融媒今日要闻',
            state: 9,
            latest_risk_level: 'high',
            latest_task_id: 208,
            latest_disposition_status: 'pending'
          },
          {
            id: 502,
            orgid: 29,
            title: '已下线稿件',
            state: 8
          }
        ]
      };
    });

    renderPage();

    expect(screen.queryByText('基于现有巡检结果聚合出的稿件工作台视图。')).not.toBeInTheDocument();
    expect(screen.queryByText('查看真实稿件元数据，并结合最近巡检结果快速进入处置。')).not.toBeInTheDocument();
    expect(screen.queryByText('支持按标题关键词筛选已发布文稿。')).not.toBeInTheDocument();
    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByText('已下线稿件')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: '标题模糊查询' })).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: '按文稿ID查询' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /查询文稿/ })).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        page: 1,
        pageSize: 20,
        orgid: 29
      }));
    });
    expect(screen.getByText('501')).toBeInTheDocument();
    expect(screen.queryByText('#501')).not.toBeInTheDocument();
    expect(screen.getByText('任务 #208')).toBeInTheDocument();
    await user.click(screen.getAllByRole('link', { name: '查看详情' })[0]);
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501?return_to=%2Farticles');
    });

    await user.click(screen.getByTitle('2'));

    expect(await screen.findByText('第二页稿件')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        page: 2,
        pageSize: 20
      }));
    });

    await user.type(screen.getByRole('textbox', { name: '标题模糊查询' }), '命中');
    await user.click(screen.getByRole('button', { name: /查询文稿/ }));

    expect(await screen.findByText('搜索命中文稿')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        title: '命中'
      }));
    });

    await user.clear(screen.getByRole('textbox', { name: '标题模糊查询' }));
    await user.type(screen.getByRole('textbox', { name: '按文稿ID查询' }), '901');
    await user.click(screen.getByRole('button', { name: /查询文稿/ }));

    expect(await screen.findByText('ID 精确命中文稿')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        article_id: 901
      }));
    });
  });

  it('re-fetches the same query after loading, and blocks duplicate clicks while loading', async () => {
    const user = userEvent.setup();
    let resolveRepeatedSearch: ((value: {
      page: number;
      pageSize: number;
      total: number;
      items: Array<{ id: number; orgid: number; title: string; state: number }>;
    }) => void) | undefined;

    mockedListArticles.mockImplementation((params) => {
      if (params.title === '重复查询') {
        return new Promise((resolve) => {
          resolveRepeatedSearch = resolve;
        });
      }

      return Promise.resolve({
        page: 1,
        pageSize: 20,
        total: 1,
        items: [
          {
            id: 501,
            orgid: 29,
            title: '县域融媒今日要闻',
            state: 9
          }
        ]
      });
    });

    renderPage();

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    const titleInput = screen.getByRole('textbox', { name: '标题模糊查询' });
    const searchButton = screen.getByRole('button', { name: /查询文稿/ });
    const resetButton = screen.getByRole('button', { name: /重\s*置/ });

    await user.type(titleInput, '重复查询');
    await user.click(searchButton);

    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledTimes(2);
    });
    expect(searchButton).toBeDisabled();
    expect(resetButton).toBeDisabled();

    await user.click(searchButton);
    expect(mockedListArticles).toHaveBeenCalledTimes(2);

    resolveRepeatedSearch?.({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 902,
          orgid: 29,
          title: '重复查询结果',
          state: 8
        }
      ]
    });

    expect(await screen.findByText('重复查询结果')).toBeInTheDocument();

    await user.click(searchButton);

    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledTimes(3);
      expect(mockedListArticles).toHaveBeenLastCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        title: '重复查询'
      }));
    });
  });

  it('hydrates article filters from the URL and restores the same state after remount', async () => {
    const user = userEvent.setup();
    const firstRender = renderPage(['/articles?title=命中&page=2&article_id=901']);

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByLabelText('标题模糊查询')).toHaveValue('命中');
    expect(screen.getByLabelText('按文稿ID查询')).toHaveValue('901');
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenLastCalledWith(expect.objectContaining({
        orgid: 29,
        page: 2,
        pageSize: 20,
        title: '命中',
        article_id: 901
      }));
    });

    await user.clear(screen.getByLabelText('标题模糊查询'));
    await user.type(screen.getByLabelText('标题模糊查询'), '复盘');
    await user.clear(screen.getByLabelText('按文稿ID查询'));
    await user.type(screen.getByLabelText('按文稿ID查询'), '902');
    await user.click(screen.getByRole('button', { name: /查询文稿/ }));

    await waitFor(() => {
      const locationState = readLocationState();
      expect(locationState.pathname).toBe('/articles');
      expect(locationState.searchParams.get('title')).toBe('复盘');
      expect(locationState.searchParams.get('article_id')).toBe('902');
    });

    firstRender.unmount();
    renderPage(['/articles?title=复盘&article_id=902']);

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByLabelText('标题模糊查询')).toHaveValue('复盘');
    expect(screen.getByLabelText('按文稿ID查询')).toHaveValue('902');
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenLastCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        title: '复盘',
        article_id: 902
      }));
    });
  });
});
