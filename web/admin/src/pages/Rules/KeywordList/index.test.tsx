import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import KeywordListPage from './index';

const { mockedListKeywords, mockedListEnabledCategories } = vi.hoisted(() => ({
  mockedListKeywords: vi.fn(),
  mockedListEnabledCategories: vi.fn()
}));

vi.mock('@/services/keywords', () => ({
  listKeywords: mockedListKeywords,
  createKeyword: vi.fn(),
  updateKeyword: vi.fn(),
  patchKeywordStatus: vi.fn(),
  deleteKeyword: vi.fn()
}));

vi.mock('@/services/categories', () => ({
  listEnabledCategories: mockedListEnabledCategories
}));

describe('KeywordListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListEnabledCategories.mockResolvedValue([
      {
        id: 12,
        orgid: 29,
        name: '涉政风险',
        enabled: true,
        sort: 1
      }
    ]);
    mockedListKeywords.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 99,
          orgid: 29,
          name: '北京市',
          category_id: 12,
          category_name: '涉政风险',
          match_type: 'contains',
          risk_level: 'high',
          suggest_action: 'offline',
          enabled: true,
          remark: '重点关注',
          scopes: ['title', 'body']
        }
      ]
    });
  });

  it('renders the keyword workspace with light table surfaces and status tags', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/keywords']}>
          <KeywordListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('heading', { name: '规则管理' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新增规则' })).toBeInTheDocument();
    expect(await screen.findByText('北京市')).toBeInTheDocument();
    expect(container.querySelector('.admin-filter-card.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell .ant-tag')).toBeInTheDocument();
  });

  it('opens the keyword form modal and delete popconfirm with light overlay classes', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/keywords']}>
          <KeywordListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('北京市')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '新增规则' }));

    expect(await screen.findByRole('dialog', { name: '新增规则' })).toBeInTheDocument();
    expect(document.querySelector('.admin-light-modal.admin-keyword-modal')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '删除规则' }));

    expect(await screen.findByText('删除后将移除该规则及其扫描范围配置。')).toBeInTheDocument();
    expect(document.querySelector('.admin-light-popconfirm.admin-keyword-popconfirm')).toBeInTheDocument();
  });
});
