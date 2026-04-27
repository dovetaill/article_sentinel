import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
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

      if (params.query) {
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
          }
        ]
      };
    });

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter>
            <ArticlesPage />
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(screen.queryByText('基于现有巡检结果聚合出的稿件工作台视图。')).not.toBeInTheDocument();
    expect(screen.queryByText('查看真实稿件元数据，并结合最近巡检结果快速进入处置。')).not.toBeInTheDocument();
    expect(screen.queryByText('支持按标题关键词筛选已发布文稿。')).not.toBeInTheDocument();
    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: '搜索文稿' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '查询文稿' })).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        page: 1,
        pageSize: 20,
        orgid: 29,
        state: 9
      }));
    });
    expect(screen.getByText('#501')).toBeInTheDocument();
    expect(screen.getByText('任务 #208')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/articles/501');

    await user.click(screen.getByTitle('2'));

    expect(await screen.findByText('第二页稿件')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        page: 2,
        pageSize: 20,
        state: 9
      }));
    });

    await user.type(screen.getByRole('textbox', { name: '搜索文稿' }), '命中');
    await user.click(screen.getByRole('button', { name: '查询文稿' }));

    expect(await screen.findByText('搜索命中文稿')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        query: '命中',
        state: 9
      }));
    });
  });
});
