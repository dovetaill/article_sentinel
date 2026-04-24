import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
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
      { id: 29, name: '一县一端', cateid: 0, enabled: true, sort: 1 }
    ]);
    mockedListArticles.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
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

  it('renders real article metadata plus latest inspect enrichment', async () => {
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
    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        state: 9
      }));
    });
    expect(screen.getByText('#501')).toBeInTheDocument();
    expect(screen.getByText('任务 #208')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/articles/501');
  });
});
