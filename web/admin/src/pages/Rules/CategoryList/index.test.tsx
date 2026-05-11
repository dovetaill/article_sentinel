import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import CategoryListPage from './index';

const { mockedListCategories } = vi.hoisted(() => ({
  mockedListCategories: vi.fn()
}));

vi.mock('@/services/categories', () => ({
  listCategories: mockedListCategories,
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  patchCategoryStatus: vi.fn(),
  deleteCategory: vi.fn(),
  listEnabledCategories: vi.fn()
}));

describe('CategoryListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListCategories.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 12,
          orgid: 29,
          name: '涉政风险',
          enabled: true,
          sort: 1
        }
      ]
    });
  });

  it('renders the category workspace with light table surfaces and status tags', async () => {
    const { container } = render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/categories']}>
          <CategoryListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('heading', { name: '规则分类' })).toBeInTheDocument();
    expect(screen.getByLabelText('分类名称')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新增分类' })).toBeInTheDocument();
    expect(await screen.findByText('涉政风险')).toBeInTheDocument();
    expect(container.querySelector('.admin-filter-card.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell.admin-surface-panel')).toBeInTheDocument();
    expect(container.querySelector('.admin-table-shell .ant-tag')).toBeInTheDocument();
  });

  it('opens the category form modal with the light overlay class', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/categories']}>
          <CategoryListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('涉政风险')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '新增分类' }));

    expect(await screen.findByRole('dialog', { name: '新增分类' })).toBeInTheDocument();
    expect(document.querySelector('.admin-light-modal.admin-category-modal')).toBeInTheDocument();
  });

  it('opens the category delete popconfirm with the light overlay class', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/categories']}>
          <CategoryListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByText('涉政风险')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '删除分类' }));

    expect(await screen.findByText('删除后该分类将不可恢复。')).toBeInTheDocument();
    expect(document.querySelector('.admin-light-popconfirm.admin-category-popconfirm')).toBeInTheDocument();
  });
});
