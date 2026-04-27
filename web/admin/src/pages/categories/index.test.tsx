import { ConfigProvider } from 'antd';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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
import { listOrgs } from '../../services/orgs';

vi.mock('../../services/categories', () => ({
  listCategories: vi.fn(),
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  patchCategoryStatus: vi.fn(),
  deleteCategory: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
}));

const mockedListCategories = vi.mocked(listCategories);
const mockedCreateCategory = vi.mocked(createCategory);
const mockedUpdateCategory = vi.mocked(updateCategory);
const mockedPatchCategoryStatus = vi.mocked(patchCategoryStatus);
const mockedDeleteCategory = vi.mocked(deleteCategory);
const mockedListOrgs = vi.mocked(listOrgs);

function renderPage() {
  return render(
    <ConfigProvider>
      <OrgProvider>
        <CategoriesPage />
      </OrgProvider>
    </ConfigProvider>,
  );
}

describe('CategoriesPage', () => {
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

    mockedListCategories.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 501,
          orgid: 29,
          name: '政策红线',
          code: 'policy-risk',
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
      code: 'freq-risk',
      enabled: true,
      sort: 20
    } as never);

    mockedUpdateCategory.mockResolvedValue({
      id: 501,
      orgid: 29,
      name: '政策调整',
      code: 'policy-risk',
      enabled: true,
      sort: 10
    } as never);

    mockedPatchCategoryStatus.mockResolvedValue({
      id: 501,
      orgid: 29,
      name: '政策红线',
      code: 'policy-risk',
      enabled: false,
      sort: 10
    } as never);

    mockedDeleteCategory.mockResolvedValue({ id: 501 });
  });

  it('loads categories for the current org and wires create/edit modal submissions', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('政策红线')).toBeInTheDocument();
    expect(screen.getByText('一县一端')).toBeInTheDocument();
    expect(screen.queryByText('先维护规则分类，再到规则管理里把具体规则挂到分类下。')).not.toBeInTheDocument();
    expect(screen.queryByText('当前机构：一县一端')).not.toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: '查看规则' })).toHaveLength(1);
    expect(screen.queryByText('按机构维护关键词规则分类，统一控制启停和排序。')).not.toBeInTheDocument();

    await waitFor(() => {
      expect(mockedListCategories).toHaveBeenCalledWith(expect.objectContaining({ orgid: 29 }));
    });

    await user.click(screen.getByRole('button', { name: '新增分类' }));
    const createDialog = await screen.findByRole('dialog', { name: '新增分类' });
    expect(createDialog).toBeInTheDocument();

    await user.type(within(createDialog).getByLabelText('分类名称'), '高频违规');
    await user.type(within(createDialog).getByLabelText('分类编码'), 'freq-risk');
    await user.click(within(createDialog).getByRole('button', { name: '确认新增' }));

    await waitFor(() => {
      expect(mockedCreateCategory).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        name: '高频违规',
        code: 'freq-risk'
      }));
    });

    await user.click(screen.getByRole('button', { name: '编辑分类' }));
    const editDialog = await screen.findByRole('dialog', { name: '编辑分类' });
    expect(editDialog).toBeInTheDocument();

    await user.clear(within(editDialog).getByLabelText('分类名称'));
    await user.type(within(editDialog).getByLabelText('分类名称'), '政策调整');
    await user.click(within(editDialog).getByRole('button', { name: '保存修改' }));

    await waitFor(() => {
      expect(mockedUpdateCategory).toHaveBeenCalledWith(501, expect.objectContaining({
        orgid: 29,
        name: '政策调整'
      }));
    });
  }, 10_000);

  it('wires status toggles and delete actions through the category service seam', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('政策红线')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '停用' }));
    await waitFor(() => {
      expect(mockedPatchCategoryStatus).toHaveBeenCalledWith(501, 29, false);
    });

    await user.click(screen.getByRole('button', { name: '删除分类' }));
    await user.click(await screen.findByRole('button', { name: '确认删除' }));

    await waitFor(() => {
      expect(mockedDeleteCategory).toHaveBeenCalledWith(501, 29);
    });
  }, 10_000);
});
