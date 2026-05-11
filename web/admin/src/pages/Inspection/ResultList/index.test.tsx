import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import ResultListPage from './index';

const { mockedListResults } = vi.hoisted(() => ({
  mockedListResults: vi.fn()
}));

vi.mock('@/services/results', () => ({
  listResults: mockedListResults,
  batchOfflineResults: vi.fn(),
  getResultDetail: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('ResultListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 29,
          task_id: 208,
          article_id: 501,
          article_title: '县域融媒今日要闻',
          risk_level: 'high',
          disposition_status: 'pending',
          hit_count: 1,
          snippet: '命中片段'
        }
      ]
    });
  });

  it('renders the global result list batch action', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/results']}>
          <ResultListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();
    expect(container.querySelectorAll('.admin-summary-card.admin-surface-panel')).toHaveLength(4);
    expect(container.querySelector('.admin-filter-card.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-result-status-tag')).toBeInTheDocument();
    expect(container.querySelector('.hit-preview.admin-surface-inline')).toBeInTheDocument();
  });

  it('passes the current page as return_to when opening article detail', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/results?task_id=208']}>
          <ResultListPage />
          <LocationProbe />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '查看详情' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent(
        '/content/articles/501?return_to=%2Finspection%2Fresults%3Ftask_id%3D208'
      );
    });
  });

  it('reads task_id from the URL when loading the result workspace', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/results?task_id=208']}>
          <ResultListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('县域融媒今日要闻')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockedListResults).toHaveBeenLastCalledWith(
        expect.objectContaining({
          page: 1,
          pageSize: 20,
          task_id: 208
        })
      );
    });
  });

  it('opens the offline modal with the light result workspace class contract', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/results']}>
          <ResultListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    await screen.findByText('县域融媒今日要闻');

    await user.click(screen.getByRole('button', { name: '下线处置' }));

    const dialog = await screen.findByRole('dialog', { name: '下线处置' });
    expect(dialog).toBeInTheDocument();
    expect(document.querySelector('.admin-light-modal.admin-result-confirm-modal')).toBeInTheDocument();
    expect(within(dialog).getByText('确认对 1 篇文章执行下线处置？')).toBeInTheDocument();
  });
});
