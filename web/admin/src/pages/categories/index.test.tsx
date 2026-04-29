import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import CategoriesPage from './index';
import {
  createCategory,
  deleteCategory,
  listCategories,
  patchCategoryStatus,
  updateCategory
} from '../../services/categories';
import { WorkbenchProvider } from '../../workbench/provider';

vi.mock('../../services/categories', () => ({
  listCategories: vi.fn(),
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  patchCategoryStatus: vi.fn(),
  deleteCategory: vi.fn()
}));

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
}));

const mockedListCategories = vi.mocked(listCategories);
const mockedCreateCategory = vi.mocked(createCategory);
const mockedUpdateCategory = vi.mocked(updateCategory);
const mockedPatchCategoryStatus = vi.mocked(patchCategoryStatus);
const mockedDeleteCategory = vi.mocked(deleteCategory);

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/rules/categories']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <CategoriesPage />
            <LocationProbe />
          </WorkbenchProvider>
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

describe('CategoriesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    mockedListCategories.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 501,
          orgid: 29,
          name: '政策红线',
          enabled: true,
          sort: 10,
          creator_name: 'alice',
          updater_name: 'alice'
        }
      ]
    });

    mockedCreateCategory.mockResolvedValue({
      id: 502,
      orgid: 29,
      name: '高频违规',
      enabled: true,
      sort: 20
    } as never);

    mockedUpdateCategory.mockResolvedValue({
      id: 501,
      orgid: 29,
      name: '政策调整',
      enabled: true,
      sort: 10
    } as never);

    mockedPatchCategoryStatus.mockResolvedValue({
      id: 501,
      orgid: 29,
      name: '政策红线',
      enabled: false,
      sort: 10
    } as never);

    mockedDeleteCategory.mockResolvedValue({ id: 501 });
  });

  function expectNoOrgID(input: unknown) {
    expect(input).not.toHaveProperty('orgid');
  }

  function expectLastCategoryList(expected: Record<string, unknown>) {
    const lastCall = mockedListCategories.mock.calls.at(-1)?.[0];
    expect(lastCall).toMatchObject(expected);
    expectNoOrgID(lastCall);
  }

  it('loads categories for the current org and wires create/edit modal submissions', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('政策红线')).toBeInTheDocument();
    expect(screen.getByText('一县一端')).toBeInTheDocument();
    expect(screen.queryByText('先维护规则分类，再到规则管理里把具体规则挂到分类下。')).not.toBeInTheDocument();
    expect(screen.queryByText('当前机构：一县一端')).not.toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: '查看规则' })).toHaveLength(1);
    expect(screen.queryByText('按机构维护关键词规则分类，统一控制启停和排序。')).not.toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: '查看规则' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/rules/keywords?category_id=501');
    });

    await waitFor(() => {
      expectLastCategoryList({});
    });

    await user.click(screen.getByRole('button', { name: '新增分类' }));
    const createDialog = await screen.findByRole('dialog', { name: '新增分类' });
    expect(createDialog).toBeInTheDocument();
    expect(within(createDialog).queryByLabelText('分类编码')).not.toBeInTheDocument();

    await user.type(within(createDialog).getByLabelText('分类名称'), '高频违规');
    await user.clear(within(createDialog).getByRole('spinbutton', { name: '排序' }));
    await user.type(within(createDialog).getByRole('spinbutton', { name: '排序' }), '20');
    await user.click(within(createDialog).getByRole('button', { name: '确认新增' }));

    await waitFor(() => {
      const createInput = mockedCreateCategory.mock.calls.at(-1)?.[0];
      expect(createInput).toMatchObject({
        name: '高频违规',
        enabled: true,
        sort: 20
      });
      expectNoOrgID(createInput);
      expect(mockedCreateCategory).not.toHaveBeenCalledWith(expect.objectContaining({
        code: expect.anything()
      }));
    });

    await user.click(screen.getByRole('button', { name: '编辑分类' }));
    const editDialog = await screen.findByRole('dialog', { name: '编辑分类' });
    expect(editDialog).toBeInTheDocument();

    await user.clear(within(editDialog).getByLabelText('分类名称'));
    await user.type(within(editDialog).getByLabelText('分类名称'), '政策调整');
    await user.click(within(editDialog).getByRole('button', { name: '保存修改' }));

    await waitFor(() => {
      const updateInput = mockedUpdateCategory.mock.calls.at(-1)?.[1];
      expect(updateInput).toMatchObject({
        name: '政策调整'
      });
      expectNoOrgID(updateInput);
    });
  }, 10_000);

  it('wires status toggles and delete actions through the category service seam', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('政策红线')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '停用' }));
    await waitFor(() => {
      expect(mockedPatchCategoryStatus).toHaveBeenCalledWith(501, false);
    });

    await user.click(screen.getByRole('button', { name: '删除分类' }));
    await user.click(await screen.findByRole('button', { name: '确认删除' }));

    await waitFor(() => {
      expect(mockedDeleteCategory).toHaveBeenCalledWith(501);
    });
  }, 10_000);

  it('hydrates category filters from the URL and restores the same state after remount', async () => {
    const user = userEvent.setup();
    const firstRender = renderPage(['/rules/categories?name=政策&enabled=true&page=2']);

    expect(await screen.findByText('政策红线')).toBeInTheDocument();
    expect(screen.getByLabelText('分类名称')).toHaveValue('政策');
    await waitFor(() => {
      expectLastCategoryList({
        page: 2,
        pageSize: 20,
        name: '政策',
        enabled: true
      });
    });

    await user.clear(screen.getByLabelText('分类名称'));
    await user.type(screen.getByLabelText('分类名称'), '高频');
    await user.click(screen.getByRole('button', { name: /查\s*询/ }));

    await waitFor(() => {
      const locationState = readLocationState();
      expect(locationState.pathname).toBe('/rules/categories');
      expect(locationState.searchParams.get('name')).toBe('高频');
      expect(locationState.searchParams.get('enabled')).toBe('true');
    });

    firstRender.unmount();
    renderPage(['/rules/categories?name=高频&enabled=true']);

    expect(await screen.findByText('政策红线')).toBeInTheDocument();
    expect(screen.getByLabelText('分类名称')).toHaveValue('高频');
    await waitFor(() => {
      expectLastCategoryList({
        page: 1,
        pageSize: 20,
        name: '高频',
        enabled: true
      });
    });
  }, 10_000);
});
