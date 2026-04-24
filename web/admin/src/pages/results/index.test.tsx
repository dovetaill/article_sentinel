import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listOrgs } from '../../services/orgs';
import ResultsPage from './index';
import { batchOfflineResults, listResults } from '../../services/results';

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
}));

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedListOrgs = vi.mocked(listOrgs);
const mockedListResults = vi.mocked(listResults);
const mockedBatchOfflineResults = vi.mocked(batchOfflineResults);

describe('ResultsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      {
        id: 29,
        name: '一县一端',
        cateid: 0,
        enabled: true,
        sort: 1
      }
    ]);
    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 29,
          task_id: 77,
          article_id: 501,
          article_title: 'Spam alert',
          risk_level: 'high',
          suggest_action: 'offline',
          disposition_status: 'pending',
          hit_count: 3,
          snippet: 'This spam alert keeps repeating',
          matched_keyword: 'spam'
        }
      ]
    } as never);
    mockedBatchOfflineResults.mockResolvedValue({ action_no: 'offline-11' } as never);
  });

  it('supports row selection and batch action confirmation', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/results']}>
          <OrgProvider>
            <ResultsPage />
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();
    expect(screen.queryByText('集中查看命中文稿、风险等级与处置状态，按批次完成研判、下线与整改。')).not.toBeInTheDocument();
    expect(screen.queryByText('按当前筛选条件查看命中文稿并执行批量处置。')).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Spam alert' })).toHaveAttribute(
      'href',
      '/articles/501?return_to=%2Fresults',
    );
    expect(screen.getByRole('button', { name: '批量下线处置' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '本页全选' }));
    expect(screen.getByText('已选 1 项')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '批量下线处置' }));
    expect(await screen.findByText('确认对 1 篇文章执行下线处置？')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '确认处置' }));

    await waitFor(() => {
      expect(mockedBatchOfflineResults).toHaveBeenCalledWith({
        orgid: 29,
        result_ids: [11],
        reason: 'manual batch offline'
      });
    });
  });
});
