import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listEnabledCategories } from '../../services/categories';
import { listOrgs } from '../../services/orgs';
import KeywordsPage from './index';
import { createKeyword, listKeywords, updateKeyword } from '../../services/keywords';

vi.mock('../../services/keywords', () => ({
  listKeywords: vi.fn(),
  createKeyword: vi.fn(),
  updateKeyword: vi.fn(),
  patchKeywordStatus: vi.fn(),
  deleteKeyword: vi.fn()
}));

vi.mock('../../services/categories', () => ({
  listEnabledCategories: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
}));

const mockedListKeywords = vi.mocked(listKeywords);
const mockedCreateKeyword = vi.mocked(createKeyword);
const mockedUpdateKeyword = vi.mocked(updateKeyword);
const mockedListEnabledCategories = vi.mocked(listEnabledCategories);
const mockedListOrgs = vi.mocked(listOrgs);

describe('KeywordsPage', () => {
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

    mockedListEnabledCategories.mockResolvedValue([
      {
        id: 501,
        orgid: 29,
        name: '政策红线',
        code: 'policy-risk',
        enabled: true,
        sort: 10
      },
      {
        id: 502,
        orgid: 29,
        name: '高频违规',
        code: 'freq-risk',
        enabled: true,
        sort: 20
      }
    ]);

    mockedListKeywords.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 7,
          orgid: 29,
          name: 'spam',
          category_id: 501,
          category_name: '政策红线',
          match_type: 'contains',
          risk_level: 'high',
          suggest_action: 'offline',
          enabled: true,
          remark: 'watch closely',
          scopes: ['title', 'body']
        }
      ]
    });

    mockedCreateKeyword.mockResolvedValue({ id: 8, orgid: 29, name: 'new keyword' } as never);
    mockedUpdateKeyword.mockResolvedValue({ id: 7, orgid: 29, name: 'spam-updated' } as never);
  });

  it('renders category selection as a searchable select and submits category_id instead of raw text', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/keywords']}>
          <OrgProvider>
            <KeywordsPage />
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByText('spam')).toBeInTheDocument();
    expect(screen.getByText('政策红线')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockedListKeywords).toHaveBeenCalledWith(expect.objectContaining({ orgid: 29 }));
    });

    await user.click(screen.getByRole('button', { name: '新增规则' }));
    const dialog = await screen.findByRole('dialog', { name: '新增规则' });
    expect(dialog).toBeInTheDocument();

    expect(within(dialog).getByRole('combobox', { name: '规则分类' })).toBeInTheDocument();
    expect(within(dialog).queryByRole('textbox', { name: '规则分类' })).not.toBeInTheDocument();

    await user.type(within(dialog).getByLabelText('关键词名称'), 'new keyword');
    await user.click(within(dialog).getByRole('combobox', { name: '规则分类' }));
    await user.click(await screen.findByRole('option', { name: '政策红线' }));
    await user.click(within(dialog).getByRole('button', { name: '确认新增' }));

    await waitFor(() => {
      expect(mockedCreateKeyword).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        category_id: 501,
        name: 'new keyword'
      }));
    });
  }, 10_000);
});
