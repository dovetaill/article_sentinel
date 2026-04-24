import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import App from './App';
import { appRoutes } from './routes';

describe('App shell', () => {
  it('locks the redesigned shell navigation and header expectations', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole('navigation', { name: /主导航/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /规则中心/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /文稿中心/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /风险结果/i })).not.toBeInTheDocument();
    expect(screen.queryByText('内容巡检与处置工作台')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /一县一端/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /当前用户|退出登录/i })).toBeInTheDocument();
  });

  it('treats /articles as the article center rather than the aggregated inspect-results view', () => {
    expect(appRoutes.map((route) => route.label)).toEqual([
      '规则中心',
      '检测任务',
      '文稿中心',
      '操作日志'
    ]);
    expect(appRoutes.some((route) => route.path === '/results')).toBe(false);
    expect(appRoutes.some((route) => route.path === '/articles')).toBe(true);
  });
});
