import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listOrgs } from '../../services/orgs';
import { WorkbenchProvider } from '../../workbench/provider';
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

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/results']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <ResultsPage />
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('ResultsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      {
        id: 29,
        name: '一县一端',
        cate_id: 0,
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
          preview_field_name: 'title',
          preview_keyword_text: 'spam',
          preview_matched_text: 'spam',
          preview_snippet: 'This spam alert keeps repeating',
          extra_hit_count: 2
        }
      ]
    } as never);
    mockedBatchOfflineResults.mockResolvedValue({
      action_id: 11,
      target_count: 1,
      success_count: 1,
      fail_count: 0,
      skip_count: 0,
      status: 'success',
      action_type: 'offline'
    } as never);
  });

  it('supports row selection and batch action confirmation', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();
    expect(screen.queryByText('集中查看命中文稿、风险等级与处置状态，按批次完成研判、下线与整改。')).not.toBeInTheDocument();
    expect(screen.queryByText('按当前筛选条件查看命中文稿并执行批量处置。')).not.toBeInTheDocument();
    expect(screen.getByText('文稿ID')).toBeInTheDocument();
    expect(screen.getByText('501')).toBeInTheDocument();
    expect(screen.queryByText('#501')).not.toBeInTheDocument();
    expect(screen.getByText('标题')).toBeInTheDocument();
    expect(screen.getAllByText('spam').length).toBeGreaterThan(0);
    expect(screen.getByText((_, element) => element?.textContent === 'This spam alert keeps repeating')).toBeInTheDocument();
    expect(screen.getByText('另有 2 条命中')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '批量下线处置' })).toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: 'Spam alert' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501?return_to=%2Fresults');
    });

    await user.click(screen.getByRole('link', { name: '进入整改' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501/rectify?return_to=%2Fresults&task_id=77&result_id=11');
    });

    await user.click(screen.getByRole('button', { name: '本页全选' }));
    expect(screen.getByText('已选 1 项')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '批量下线处置' }));
    expect(await screen.findByText('确认对 1 篇文章执行下线处置？')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '确认处置' }));

    await waitFor(() => {
      expect(mockedBatchOfflineResults).toHaveBeenCalledWith({
        orgid: 29,
        task_id: 77,
        result_ids: [11],
        reason: 'manual batch offline'
      });
    });
  });
});
