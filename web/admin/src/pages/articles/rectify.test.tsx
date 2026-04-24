import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import RectifyPage from './rectify';
import { getArticleRectify, rectifyArticle } from '../../services/results';

const { mockedListOrgs } = vi.hoisted(() => ({
  mockedListOrgs: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: mockedListOrgs
}));

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  batchIgnoreResults: vi.fn(),
  batchProcessResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedGetArticleRectify = vi.mocked(getArticleRectify);
const mockedRectifyArticle = vi.mocked(rectifyArticle);

describe('RectifyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cateid: 0, enabled: true, sort: 1 }
    ]);
    mockedGetArticleRectify.mockResolvedValue({
      article_id: 501,
      orgid: 29,
      title: 'Old title',
      desc: 'Old summary',
      body: '<p>Old body</p>'
    } as never);
    mockedRectifyArticle.mockResolvedValue({ article_id: 501, status: 'saved' } as never);
  });

  it('submits rectification against the active org context', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <MemoryRouter initialEntries={['/articles/501/rectify']}>
            <Routes>
              <Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
            </Routes>
          </MemoryRouter>
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(await screen.findByText(/old title/i)).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '内容整改' })).toBeInTheDocument();
    expect(screen.queryByText(/围绕当前稿件进行标题、摘要与正文修订。/)).not.toBeInTheDocument();

    await user.clear(screen.getByLabelText('整改标题'));
    await user.type(screen.getByLabelText('整改标题'), 'New title');
    await user.clear(screen.getByLabelText('整改摘要'));
    await user.type(screen.getByLabelText('整改摘要'), 'New summary');
    await user.click(screen.getByRole('button', { name: '保存整改' }));

    await waitFor(() => {
      expect(mockedRectifyArticle).toHaveBeenCalledWith(501, {
        orgid: 29,
        title: 'New title',
        desc: 'New summary',
        body: '<p>Old body</p>',
        target_article_state: undefined
      });
    });
  });
});
