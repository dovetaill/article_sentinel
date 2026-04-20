import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import KeywordsPage from './index';
import { createKeyword, listKeywords, updateKeyword } from '../../services/keywords';

vi.mock('../../services/keywords', () => ({
  listKeywords: vi.fn(),
  createKeyword: vi.fn(),
  updateKeyword: vi.fn(),
  patchKeywordStatus: vi.fn(),
  deleteKeyword: vi.fn()
}));

const mockedListKeywords = vi.mocked(listKeywords);
const mockedCreateKeyword = vi.mocked(createKeyword);
const mockedUpdateKeyword = vi.mocked(updateKeyword);

describe('KeywordsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListKeywords.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 7,
          orgid: 100,
          name: 'spam',
          category: 'policy',
          match_type: 'contains',
          risk_level: 'high',
          suggest_action: 'offline',
          enabled: true,
          remark: 'watch closely',
          scopes: ['title', 'body']
        }
      ]
    });
    mockedCreateKeyword.mockResolvedValue({ id: 8, orgid: 100, name: 'new keyword' } as never);
    mockedUpdateKeyword.mockResolvedValue({ id: 7, orgid: 100, name: 'spam-updated' } as never);
  });

  it('renders keyword data and opens create/edit modal', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <KeywordsPage />
      </ConfigProvider>,
    );

    expect(await screen.findByText('spam')).toBeInTheDocument();
    expect(screen.getByText('policy')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /new keyword/i }));
    expect(await screen.findByRole('dialog', { name: /create keyword/i })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: /create keyword/i })).not.toBeInTheDocument();
    });

    await user.click(screen.getByRole('button', { name: /edit/i }));
    expect(await screen.findByRole('dialog', { name: /edit keyword/i })).toBeInTheDocument();
  });
});
