import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import KeywordListPage from './index';

describe('KeywordListPage', () => {
  it('renders the keyword page heading and add button', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/rules/keywords']}>
          <KeywordListPage />
        </MemoryRouter>
      </ConfigProvider>
    );

    expect(await screen.findByRole('heading', { name: '规则管理' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新增规则' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: '规则分类' })).toBeInTheDocument();
  });
});
