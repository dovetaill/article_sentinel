import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import RectifyPage from './rectify';
import { getArticleRectify, rectifyArticle } from '../../services/results';

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedGetArticleRectify = vi.mocked(getArticleRectify);
const mockedRectifyArticle = vi.mocked(rectifyArticle);

describe('RectifyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetArticleRectify.mockResolvedValue({
      article_id: 501,
      orgid: 100,
      title: 'Old title',
      desc: 'Old summary',
      body: '<p>Old body</p>'
    } as never);
    mockedRectifyArticle.mockResolvedValue({ article_id: 501, status: 'saved' } as never);
  });

  it('displays old/new values and submits update', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/articles/501/rectify']}>
          <Routes>
            <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
          </Routes>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();
    expect(screen.getByText(/old summary/i)).toBeInTheDocument();

    await user.clear(screen.getByLabelText(/new title/i));
    await user.type(screen.getByLabelText(/new title/i), 'New title');
    await user.clear(screen.getByLabelText(/new summary/i));
    await user.type(screen.getByLabelText(/new summary/i), 'New summary');
    await user.click(screen.getByRole('button', { name: /save rectification/i }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenCalledWith(501, {
        orgid: 100,
        title: 'New title',
        desc: 'New summary',
        body: '<p>Old body</p>',
        target_article_state: undefined
      });
    });
  });
});
