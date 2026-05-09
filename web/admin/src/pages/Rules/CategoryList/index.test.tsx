import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import CategoryListPage from './index';

describe('CategoryListPage', () => {
  it('renders the category page heading and query controls', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/categories']}>
          <CategoryListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('heading', { name: '规则分类' })).toBeInTheDocument();
    expect(screen.getByLabelText('分类名称')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新增分类' })).toBeInTheDocument();
  });
});
