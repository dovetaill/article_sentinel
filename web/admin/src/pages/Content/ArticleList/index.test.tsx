import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ArticleListPage from './index';

const { mockedListArticles } = vi.hoisted(() => ({
  mockedListArticles: vi.fn()
}));

vi.mock('@/services/articles', () => ({
  listArticles: mockedListArticles,
  getArticleDetail: vi.fn(),
  offlineArticle: vi.fn(),
  republishArticle: vi.fn(),
  rectifyArticle: vi.fn()
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('ArticleListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders article list filters for title and article id', async () => {
    mockedListArticles.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 0,
      items: []
    });

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/content/articles']}>
          <ArticleListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByLabelText('标题模糊查询')).toBeInTheDocument();
    expect(screen.getByLabelText('按文稿ID查询')).toBeInTheDocument();
  });

  it('re-fetches with title filters and keeps return_to when opening article detail', async () => {
    const user = userEvent.setup();

    mockedListArticles.mockImplementation(async (params?: { title?: string }) => {
      if (params?.title === '命中') {
        return {
          page: 1,
          pageSize: 20,
          total: 1,
          items: [
            {
              id: 901,
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
        total: 1,
        items: [
          {
            id: 501,
            orgid: 29,
            title: '县域融媒今日要闻',
            state: 9
          }
        ]
      };
    });

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/content/articles']}>
          <ArticleListPage />
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    await user.type(screen.getByRole('textbox', { name: '标题模糊查询' }), '命中');
    await user.click(screen.getByRole('button', { name: '查询文稿' }));

    expect(await screen.findByText('搜索命中文稿')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenLastCalledWith(
        expect.objectContaining({
          page: 1,
          pageSize: 20,
          title: '命中'
        })
      );
    });

    await user.click(screen.getByRole('button', { name: '查看详情' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent(
        '/content/articles/901?return_to=%2Fcontent%2Farticles%3Ftitle%3D%25E5%2591%25BD%25E4%25B8%25AD'
      );
    });
  });
});
