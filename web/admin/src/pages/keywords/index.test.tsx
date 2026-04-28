import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listEnabledCategories } from '../../services/categories';
import { listOrgs } from '../../services/orgs';
import KeywordsPage from './index';
import { createKeyword, deleteKeyword, listKeywords, updateKeyword } from '../../services/keywords';

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
const mockedDeleteKeyword = vi.mocked(deleteKeyword);
const mockedListEnabledCategories = vi.mocked(listEnabledCategories);
const mockedListOrgs = vi.mocked(listOrgs);

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/rules/keywords']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <KeywordsPage />
          <LocationProbe />
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function readLocationState() {
  const text = screen.getByTestId('location-probe').textContent ?? '';
  const [pathname, search = ''] = text.split('?');

  return {
    pathname,
    searchParams: new URLSearchParams(search)
  };
}

describe('KeywordsPage', () => {
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

    mockedListEnabledCategories.mockResolvedValue([
      {
        id: 501,
        orgid: 29,
        name: '政策红线',
        enabled: true,
        sort: 10
      },
      {
        id: 502,
        orgid: 29,
        name: '高频违规',
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
    mockedDeleteKeyword.mockResolvedValue({ id: 7 } as never);
  });

  it('renders category selection as a searchable select and submits category_id instead of raw text', async () => {
    const user = userEvent.setup();

    renderPage(['/rules/keywords']);

    expect(await screen.findByText('spam')).toBeInTheDocument();
    expect(screen.getByText('政策红线')).toBeInTheDocument();
    expect(screen.queryByText('新建规则时先选择分类；执行检测任务时再按规则进行勾选。')).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: '查看分类' })).not.toBeInTheDocument();

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

  it('deletes a keyword from the list after confirmation', async () => {
    const user = userEvent.setup();

    renderPage(['/rules/keywords']);

    expect(await screen.findByText('spam')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '删除规则' }));
    await user.click(await screen.findByRole('button', { name: '确认删除' }));

    await waitFor(() => {
      expect(mockedDeleteKeyword).toHaveBeenCalledWith(7, 29);
    });
  });

  it('hydrates keyword filters from the URL and restores the same state after remount', async () => {
    const user = userEvent.setup();
    const firstRender = renderPage(['/rules/keywords?name=spam&category_id=501&page=2']);

    expect(await screen.findByText('spam')).toBeInTheDocument();
    expect(screen.getByLabelText('关键词名称')).toHaveValue('spam');
    await waitFor(() => {
      expect(mockedListKeywords).toHaveBeenLastCalledWith(expect.objectContaining({
        orgid: 29,
        page: 2,
        pageSize: 20,
        keyword: 'spam',
        categoryId: 501
      }));
    });

    await user.clear(screen.getByLabelText('关键词名称'));
    await user.type(screen.getByLabelText('关键词名称'), 'risk');
    await user.click(screen.getByRole('button', { name: /查\s*询/ }));

    await waitFor(() => {
      const locationState = readLocationState();
      expect(locationState.pathname).toBe('/rules/keywords');
      expect(locationState.searchParams.get('name')).toBe('risk');
      expect(locationState.searchParams.get('category_id')).toBe('501');
    });

    firstRender.unmount();
    renderPage(['/rules/keywords?name=risk&category_id=501']);

    expect(await screen.findByText('spam')).toBeInTheDocument();
    expect(screen.getByLabelText('关键词名称')).toHaveValue('risk');
    await waitFor(() => {
      expect(mockedListKeywords).toHaveBeenLastCalledWith(expect.objectContaining({
        orgid: 29,
        page: 1,
        pageSize: 20,
        keyword: 'risk',
        categoryId: 501
      }));
    });
  });
});
