import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import App from './App';
import { appRoutes } from './routes';

describe('App shell', () => {
  it('renders a compact dashboard shell without the oversized hero metadata cards', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getAllByText('融媒内容安全巡检平台')).toHaveLength(1);
    expect(screen.getByRole('heading', { name: '关键词规则' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /关键词规则/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /风险结果/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /文稿列表/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
    expect(screen.queryByText('适用机构')).not.toBeInTheDocument();
    expect(screen.queryByText('巡检时段')).not.toBeInTheDocument();
    expect(screen.queryByText('提示状态')).not.toBeInTheDocument();
    expect(screen.queryByText('当前环境')).not.toBeInTheDocument();
    expect(screen.queryByText('值守模式')).not.toBeInTheDocument();
    expect(screen.queryByText('政务融媒')).not.toBeInTheDocument();
    expect(screen.queryByText('值守中')).not.toBeInTheDocument();
    expect(screen.queryByText('控制台')).not.toBeInTheDocument();
    expect(screen.queryByText('巡检控制台')).not.toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: /article sentinel/i })).not.toBeInTheDocument();
    expect(screen.queryByText('安全巡检后台')).not.toBeInTheDocument();
    expect(screen.queryByText('留存任务执行与稿件处置过程的关键记录。')).not.toBeInTheDocument();
  });

  it('uses Chinese labels for the primary navigation routes', () => {
    expect(appRoutes.map((route) => route.label)).toEqual([
      '关键词规则',
      '检测任务',
      '风险结果',
      '文稿列表',
      '操作日志'
    ]);
  });

  it('resolves article list and article detail as non-sidebar routes', () => {
    render(
      <MemoryRouter initialEntries={['/articles']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getAllByRole('heading', { name: '文稿列表' }).length).toBeGreaterThan(0);

    cleanup();

    render(
      <MemoryRouter initialEntries={['/articles/501']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getAllByRole('heading', { name: '文稿详情' }).length).toBeGreaterThan(0);
  });
});
