import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import App from './App';
import { appRoutes } from './routes';

describe('App shell', () => {
  it('renders the formal Chinese platform shell', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getAllByText('融媒内容安全巡检平台').length).toBeGreaterThan(0);
    expect(screen.getByRole('heading', { name: '关键词规则' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /关键词规则/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /检测任务/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /风险结果/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /操作日志/i })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: /article sentinel/i })).not.toBeInTheDocument();
    expect(screen.queryByText('安全巡检后台')).not.toBeInTheDocument();
  });

  it('uses Chinese labels for the primary navigation routes', () => {
    expect(appRoutes.map((route) => route.label)).toEqual(['关键词规则', '检测任务', '风险结果', '操作日志']);
  });
});
