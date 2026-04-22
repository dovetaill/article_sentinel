import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ResultsPage from './index';
import { batchOfflineResults, listResults } from '../../services/results';

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedListResults = vi.mocked(listResults);
const mockedBatchOfflineResults = vi.mocked(batchOfflineResults);

describe('ResultsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 100,
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
        <ResultsPage />
      </ConfigProvider>,
    );

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '风险结果' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '本页全选' }));
    expect(screen.getByText('已选 1 项')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '批量下线处置' }));
    expect(await screen.findByText('确认对 1 篇文章执行下线处置？')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '确认处置' }));

    await waitFor(() => {
      expect(mockedBatchOfflineResults).toHaveBeenCalledWith({
        orgid: 100,
        result_ids: [11],
        reason: 'manual batch offline'
      });
    });
  });
});
