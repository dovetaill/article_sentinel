import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import ResultListPage from './index';

vi.mock('@/services/results', () => ({
  listResults: vi.fn().mockResolvedValue({
    page: 1,
    pageSize: 20,
    total: 0,
    items: []
  }),
  batchOfflineResults: vi.fn(),
  getResultDetail: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

describe('ResultListPage', () => {
  it('renders the global result list batch action', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/inspection/results']}>
          <ResultListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
  });
});
